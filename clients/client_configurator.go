//go:build clients
// +build clients

// replication-manager - Replication Manager Monitoring and CLI for MariaDB and MySQL
// Copyright 2017-2021 SIGNAL18 CLOUD SAS
// Author: Stephane Varoqui  <svaroqui@gmail.com>
// License: GNU General Public License, version 3. Redistribution/Reuse of this code is permitted under the GNU v3 license, as an additional term ALL code must carry the original Author(s) credit in comment form.
// See LICENSE in this directory for the integral text.

package clients

import (
	"fmt"
	"io"
	"os"
	"runtime/pprof"
	"sort"
	"strings"
	"time"

	termbox "github.com/nsf/termbox-go"
	"github.com/signal18/replication-manager/cluster/configurator"
	v3 "github.com/signal18/replication-manager/repmanv3"
	"github.com/signal18/replication-manager/server"
	"github.com/signal18/replication-manager/utils/dbhelper"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var dbCurrrentTag string
var dbCurrrentCategory string
var dbCategories map[string]string
var dbCategoriesSortedKeys []string
var dbResourceCategories = []string{"MEMORY", "DISK", "CPU", "NETWORK"}
var dbResourceCategoryIndex int = 0
var dbUsedTags []string
var dbCategoryIndex int
var dbTagIndex int
var dbCurrentCategoryTags []v3.Tag
var dbUsedTagIndex int
var PanIndex int
var dbHost string
var dbUser string
var dbPassword string
var memoryInput string
var ioDiskInput string
var coresInput string
var connectionsInput string
var inputMode bool = false
var cursorPos int = 0
var RepMan *server.ReplicationManager
var addedTags = make(map[string]bool)

const maxTagsInView = 5          // Nombre maximum de tags à afficher à la fois
var startViewIndex = 0           // Index de début de la fenêtre de visualisation
var endViewIndex = maxTagsInView // Index de fin de la fenêtre de visualisation, initialement réglé sur maxTagsInView

var configuratorCmd = &cobra.Command{
	Use:   "configurator",
	Short: "Config generator",
	Long:  `Config generator produce tar.gz for databases and proxies based on ressource and tags description`,
	Run: func(cmd *cobra.Command, args []string) {
		conf.WithEmbed = WithEmbed
		RepMan = new(server.ReplicationManager)
		RepMan.SetDefaultFlags(viper.GetViper())
		conf.HttpServ = false
		conf.ApiServ = false
		RepMan.InitConfig(conf, false)
		go RepMan.Run()
		time.Sleep(2 * time.Second)
		cluster := RepMan.Clusters[RepMan.ClusterList[0]]

		if cluster == nil {
			log.Fatalf("No Cluster found .replication-manager/config.toml")
		}
		for _, s := range cluster.Servers {
			conn, err := s.GetNewDBConn()
			if err != nil {
				log.WithError(err).Fatalf("Connecting error to database in .replication-manager/config.toml: %s", s.URL)
			}
			variables, _, err := dbhelper.GetVariablesCase(conn, s.DBVersion, "LOWER")
			if err != nil {
				log.WithError(err).Fatalf("Get variables failed %s", s.URL)
			}

			log.Infof("datadir %s", variables["DATADIR"])
		}
		RepMan.Clusters["mycluster"].WaitDatabaseCanConn()
		//	var conf config.Config
		//var configurator configurator.Configurator
		//configurator.Init(conf)
		//configurator := cluster.Configurator

		for _, server := range cluster.Servers {
			err := cluster.Configurator.GenerateDatabaseConfig(server.Datadir, cluster.Conf.WorkingDir, server.GetVariablesCaseSensitive()["DATADIR"], server.GetEnv(), cluster.RepMgrVersion)
			if err != nil {
				log.WithError(err).Fatalf("Generate database config failed %s", server.URL)
			}
			log.Infof("Generate database config datadir %s/config.tar.gz", server.Datadir)
		}

		dbCategories = cluster.Configurator.GetDBModuleCategories()
		dbCategoriesSortedKeys = make([]string, 0, len(dbCategories))
		for k := range dbCategories {
			dbCategoriesSortedKeys = append(dbCategoriesSortedKeys, k)
		}

		sort.Strings(dbCategoriesSortedKeys)
		defaultTags := cluster.Configurator.GetDBTags()
		for _, v := range defaultTags {

			addedTags[v] = true
			//fmt.Printf("%s \n" ,v)
		}
		//os.Exit(3)

		//default
		memoryInput = cluster.Configurator.GetConfigDBMemory()
		ioDiskInput = cluster.Configurator.GetConfigDBDiskIOPS()
		coresInput = cluster.Configurator.GetConfigDBCores()
		connectionsInput = cluster.Configurator.GetConfigMaxConnections()

		fmt.Printf("%s \n", RepMan.Clusters["mycluster"].Conf.ProvTags)
		conf.SetLogOutput(io.Discard)
		err := termbox.Init()
		if err != nil {
			log.WithError(err).Fatal("Termbox initialization error")
		}
		_, cliTermlength = termbox.Size()
		if cliTermlength == 0 {
			cliTermlength = 120
		} else if cliTermlength < 18 {
			log.Fatal("Terminal too small, please increase window size")
		}
		termboxChan := cliNewTbChan()
		interval := time.Millisecond
		ticker := time.NewTicker(interval * time.Duration(20))

		cliDisplayConfigurator(&cluster.Configurator)

		for cliExit == false {
			select {
			case <-ticker.C:
				cliDisplayConfigurator(&cluster.Configurator)

			case event := <-termboxChan:
				switch event.Type {
				case termbox.EventKey:
					if event.Key == termbox.KeyCtrlS {
						cluster.Save()
						for _, server := range cluster.Servers {
							err := cluster.Configurator.GenerateDatabaseConfig(server.Datadir, cluster.Conf.WorkingDir, server.GetVariablesCaseSensitive()["DATADIR"], server.GetEnv(), cluster.RepMgrVersion)
							if err != nil {
								log.WithError(err).Fatalf("Generate database config failed %s", server.URL)
							}
							log.Infof("Generate database config datadir %s/config.tar.gz", server.Datadir)
						}
						cliExit = true
					}

					if event.Key == termbox.KeyArrowLeft {
						switch PanIndex {
						case 0:
							dbCategoryIndex--
							dbTagIndex = 0
							if dbCategoryIndex < 0 {
								dbCategoryIndex = len(dbCategoriesSortedKeys) - 1
							}
						case 2:
							dbResourceCategoryIndex--
							if dbResourceCategoryIndex < 0 {
								dbResourceCategoryIndex = len(dbResourceCategories) - 1
							}
						case 3:
							if cursorPos > 0 {
								// Déplacer le curseur vers la gauche
								cursorPos--
							}
						default:
						}
					}

					if event.Key == termbox.KeyArrowRight {
						switch PanIndex {
						case 0:
							dbCategoryIndex++
							dbTagIndex = 0
							if dbCategoryIndex >= len(dbCategoriesSortedKeys) {
								dbCategoryIndex = 0
							}
						case 2:
							dbResourceCategoryIndex++
							if dbResourceCategoryIndex >= len(dbResourceCategories) {
								dbResourceCategoryIndex = 0
							}
						case 3:
							switch dbResourceCategoryIndex {
							case 0:
								if cursorPos < len(memoryInput) {
									// Déplacer le curseur vers la droite
									cursorPos++
								}
							case 1:
								if cursorPos < len(ioDiskInput) {
									// Déplacer le curseur vers la droite
									cursorPos++
								}
							case 2:
								if cursorPos < len(coresInput) {
									// Déplacer le curseur vers la droite
									cursorPos++
								}
							case 3:
								if cursorPos < len(connectionsInput) {
									// Déplacer le curseur vers la droite
									cursorPos++
								}
							default:
							}
						default:
						}
					}

					if event.Key == termbox.KeyArrowDown {
						switch PanIndex {
						case 0:
							PanIndex = 2
						case 1:
							if dbTagIndex < len(dbCurrentCategoryTags)-1 {
								dbTagIndex++
								if dbTagIndex >= endViewIndex && endViewIndex < len(dbCurrentCategoryTags) {
									// Faites défiler la fenêtre de visualisation vers le bas
									startViewIndex++
									endViewIndex++
								}
							}
						case 2:
							PanIndex = 0
						default:
						}
					}

					if event.Key == termbox.KeyArrowUp {
						switch PanIndex {
						case 0:
							PanIndex = 2
						case 1:
							if dbTagIndex > 0 {
								dbTagIndex--
								if dbTagIndex < startViewIndex && startViewIndex > 0 {
									// Faites défiler la fenêtre de visualisation vers le haut
									startViewIndex--
									endViewIndex--
								}
							}
						case 2:
							PanIndex = 0
						default:
						}
					}

					if event.Key == termbox.KeyEnter {
						switch PanIndex {
						case 0:
							PanIndex = 1
						case 1:
							if addedTags[dbCurrrentTag] {
								cluster.DropDBTag(dbCurrrentTag)
								addedTags[dbCurrrentTag] = false
							} else {
								cluster.AddDBTag(dbCurrrentTag)
								addedTags[dbCurrrentTag] = true
							}
							cluster.SetTagsFromConfigurator()

						case 2:
							PanIndex = 3
							inputMode = true
							switch dbResourceCategoryIndex {
							case 0:
								cursorPos = len(memoryInput)
							case 1:
								cursorPos = len(ioDiskInput)
							case 2:
								cursorPos = len(coresInput)
							case 3:
								cursorPos = len(connectionsInput)
							default:
							}
						case 3:
							switch dbResourceCategoryIndex {
							case 0: // MEMORY
								cluster.SetDBMemorySize(memoryInput)
							case 1: // DISK
								cluster.SetDBDiskSize(ioDiskInput)
							case 2: // CPU
								cluster.SetDBCores(coresInput)
							case 3: // NETWORK
								// Supposons que vous avez une fonction pour régler les connections
								cluster.SetDBMaxConnections(connectionsInput)
							default:
							}
							inputMode = false
							cursorPos = 0
							PanIndex = 2
						default:
						}
					}

					if inputMode {
						// Gérer la saisie de l'utilisateur dans le mode de saisie
						if event.Ch != 0 && event.Ch >= '0' && event.Ch <= '9' { // Vérifier si le caractère est un chiffre
							// Ajouter un nouveau caractère à la position du curseur
							switch dbResourceCategoryIndex {
							case 0:
								if len(memoryInput) < 6 {
									memoryInput = memoryInput[:cursorPos] + string(event.Ch) + memoryInput[cursorPos:]
									cursorPos++
								}
							case 1:
								if len(ioDiskInput) < 6 {
									ioDiskInput = ioDiskInput[:cursorPos] + string(event.Ch) + ioDiskInput[cursorPos:]
									cursorPos++
								}
							case 2:
								if len(coresInput) < 6 {
									coresInput = coresInput[:cursorPos] + string(event.Ch) + coresInput[cursorPos:]
									cursorPos++
								}
							case 3:
								if len(connectionsInput) < 6 {
									connectionsInput = connectionsInput[:cursorPos] + string(event.Ch) + connectionsInput[cursorPos:]
									cursorPos++
								}
							default:
							}
						}
					}
					if event.Key == termbox.KeyBackspace || event.Key == termbox.KeyBackspace2 {
						switch dbResourceCategoryIndex {
						case 0:
							if cursorPos > 0 && len(memoryInput) > 0 {
								// Supprimer le caractère à gauche du curseur
								memoryInput = memoryInput[:cursorPos-1] + memoryInput[cursorPos:]
								cursorPos--
							}
						case 1:
							if cursorPos > 0 && len(ioDiskInput) > 0 {
								// Supprimer le caractère à gauche du curseur
								ioDiskInput = ioDiskInput[:cursorPos-1] + ioDiskInput[cursorPos:]
								cursorPos--
							}
						case 2:
							if cursorPos > 0 && len(coresInput) > 0 {
								// Supprimer le caractère à gauche du curseur
								coresInput = coresInput[:cursorPos-1] + coresInput[cursorPos:]
								cursorPos--
							}
						case 3:
							if cursorPos > 0 && len(connectionsInput) > 0 {
								// Supprimer le caractère à gauche du curseur
								connectionsInput = connectionsInput[:cursorPos-1] + connectionsInput[cursorPos:]
								cursorPos--
							}
						default:
						}
					}

					if event.Key == termbox.KeyEsc {
						switch PanIndex {
						case 1:
							dbTagIndex = 0
							startViewIndex = 0
							endViewIndex = maxTagsInView
							PanIndex = 0
						case 3:
							switch dbResourceCategoryIndex {
							case 0: // MEMORY
								cluster.SetDBMemorySize(memoryInput)
							case 1: // DISK
								cluster.SetDBDiskSize(ioDiskInput)
							case 2: // CPU
								cluster.SetDBCores(coresInput)
							case 3: // NETWORK
								// Supposons que vous avez une fonction pour régler les connections
								cluster.SetDBMaxConnections(connectionsInput)
							default:
							}
							inputMode = false
							cursorPos = 0
							PanIndex = 2
						default:
						}
					}

					if event.Key == termbox.KeyCtrlH {
						cliDisplayHelp()
					}
					if event.Key == termbox.KeyCtrlQ {
						cliExit = true
					}
					if event.Key == termbox.KeyCtrlC {
						cliExit = true
					}

				}
				switch event.Ch {
				//	case 's':
				//		termbox.Sync()
				default:
				}
				cliDisplayConfigurator(&cluster.Configurator)

			}
		}
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		// Close connections on exit.
		termbox.Close()
		if memprofile != "" {
			f, err := os.Create(memprofile)
			if err != nil {
				log.Fatal(err)
			}
			pprof.WriteHeapProfile(f)
			f.Close()
		}
		RepMan.Stop()
	},
}

func cliDisplayConfigurator(configurator *configurator.Configurator) {

	termbox.Clear(termbox.ColorWhite, termbox.ColorBlack)
	headstr := fmt.Sprintf(" Signal18 Replication Manager Configurator")

	cliPrintfTb(0, 0, termbox.ColorWhite, termbox.ColorBlack|termbox.AttrReverse|termbox.AttrBold, headstr)
	cliPrintfTb(0, 1, termbox.ColorRed, termbox.ColorBlack|termbox.AttrReverse|termbox.AttrBold, cliConfirm)
	cliTlog.Line = 3
	tableau := "─"
	tags := configurator.GetDBModuleTags()
	width, _ := termbox.Size()

	//PanIndex = 0 -- TAGS
	colorCell := termbox.ColorWhite
	if PanIndex == 0 {
		colorCell = termbox.ColorCyan
	} else {
		colorCell = termbox.ColorWhite
	}

	cliPrintfTb(0, cliTlog.Line, colorCell|termbox.AttrBold, termbox.ColorBlack, "%s", strings.Repeat(tableau, width))
	cliTlog.Line++
	cliPrintTb(1, cliTlog.Line, termbox.ColorWhite, termbox.ColorBlack, "DATABASE FEATURES :")
	curWitdh := len("DATABASE FEATURES :") + 2

	for i, cat := range dbCategoriesSortedKeys {
		tag := dbCategories[cat]

		if dbCurrrentCategory == "" || i == dbCategoryIndex {
			dbCurrrentCategory = cat
			if dbCurrrentTag == "" {
				dbCurrrentTag = tag
			}
		}

		if curWitdh > width {
			curWitdh = 1
			cliTlog.Line++
		}
		if dbCurrrentCategory != cat {
			cliPrintTb(curWitdh, cliTlog.Line, termbox.ColorWhite, termbox.ColorBlack, strings.ToUpper(cat))
		} else {
			cliPrintTb(curWitdh, cliTlog.Line, termbox.ColorBlack, colorCell, strings.ToUpper(cat))
		}
		curWitdh += len(cat)
		curWitdh++
	}

	cliTlog.Line++
	cliPrintfTb(0, cliTlog.Line, colorCell|termbox.AttrBold, termbox.ColorBlack, "%s", strings.Repeat(tableau, width))
	cliTlog.Line++
	cliTlog.Line++

	//PanIndex = 1 -- print available tags for a category
	if PanIndex == 1 {
		colorCell = termbox.ColorCyan
	} else {
		colorCell = termbox.ColorWhite
	}

	curWitdh = 1

	dbCurrentCategoryTags = make([]v3.Tag, 0, len(tags))
	dbUsedTags = configurator.GetDBTags()

	for _, tag := range tags {
		if dbCurrrentCategory == tag.Category /*&& !configurator.HaveDBTag(tag.Name)*/ {
			dbCurrentCategoryTags = append(dbCurrentCategoryTags, tag)
		}
	}

	for i := startViewIndex; i < endViewIndex && i < len(dbCurrentCategoryTags); i++ {
		tag := dbCurrentCategoryTags[i]
		var tagDisplay string
		if addedTags[tag.Name] {
			tagDisplay = "[X] " + tag.Name
		} else {
			tagDisplay = "[ ] " + tag.Name
		}
		if i == dbTagIndex && PanIndex == 1 {
			cliPrintTb(curWitdh, cliTlog.Line, termbox.ColorBlack, colorCell, tagDisplay)
			dbCurrrentTag = tag.Name
		} else {
			cliPrintTb(curWitdh, cliTlog.Line, termbox.ColorWhite, termbox.ColorBlack, tagDisplay)
		}
		cliTlog.Line++
	}

	//PanIndex 2 ou plus
	if PanIndex == 2 || PanIndex == 3 || PanIndex == 4 || PanIndex == 5 {
		colorCell = termbox.ColorCyan
	} else {
		colorCell = termbox.ColorWhite
	}
	curWitdh = len("OS RESSOURCES :") + 2
	cliTlog.Line++
	cliPrintfTb(0, cliTlog.Line, colorCell|termbox.AttrBold, termbox.ColorBlack, "%s", strings.Repeat(tableau, width))
	cliTlog.Line++
	cliPrintTb(1, cliTlog.Line, termbox.ColorWhite, termbox.ColorBlack, "OS RESSOURCES :")
	for i, res := range dbResourceCategories {
		if i == dbResourceCategoryIndex {
			cliPrintTb(curWitdh, cliTlog.Line, termbox.ColorBlack, colorCell, res)
		} else {
			cliPrintTb(curWitdh, cliTlog.Line, termbox.ColorWhite, termbox.ColorBlack, res)
		}

		curWitdh += len(res)
		curWitdh++
	}

	cliTlog.Line++
	cliPrintfTb(0, cliTlog.Line, colorCell|termbox.AttrBold, termbox.ColorBlack, "%s", strings.Repeat(tableau, width))
	cliTlog.Line++
	cliTlog.Line++

	// Déterminez la largeur maximale des étiquettes
	labels := []string{"MEMORY :", "IO DISK :", "CORES :", "CONNECTIONS :"}
	maxLabelWidth := 0
	for _, label := range labels {
		if len(label) > maxLabelWidth {
			maxLabelWidth = len(label)
		}
	}

	if PanIndex == 2 {
		colorCell = termbox.ColorCyan
	} else {
		colorCell = termbox.ColorWhite
	}

	switch dbResourceCategoryIndex {
	case 0:
		if !inputMode {
			formattedInput := formatInput(memoryInput, cursorPos, inputMode)
			formattedInput = formattedInput + " Mb"
			cliPrintTb(1, cliTlog.Line, termbox.ColorWhite, termbox.ColorBlack, formattedInput)
		} else {
			formattedInput := formatInput(memoryInput, cursorPos, inputMode)
			formattedInput = formattedInput + " Mb"
			cliPrintTb(1, cliTlog.Line, colorCell|termbox.AttrBold, termbox.ColorBlack, formattedInput)
		}
	case 1:
		if !inputMode {
			formattedInput := formatInput(ioDiskInput, cursorPos, inputMode)
			formattedInput = formattedInput + " Random iops"
			cliPrintTb(1, cliTlog.Line, termbox.ColorWhite, termbox.ColorBlack, formattedInput)
		} else {
			formattedInput := formatInput(ioDiskInput, cursorPos, inputMode)
			formattedInput = formattedInput + " Random iops"
			cliPrintTb(1, cliTlog.Line, colorCell|termbox.AttrBold, termbox.ColorBlack, formattedInput)
		}
	case 2:
		if !inputMode {
			formattedInput := formatInput(coresInput, cursorPos, inputMode)
			formattedInput = formattedInput + " Cores"
			cliPrintTb(1, cliTlog.Line, termbox.ColorWhite, termbox.ColorBlack, formattedInput)
		} else {
			formattedInput := formatInput(coresInput, cursorPos, inputMode)
			formattedInput = formattedInput + " Cores"
			cliPrintTb(1, cliTlog.Line, colorCell|termbox.AttrBold, termbox.ColorBlack, formattedInput)
		}
	case 3:
		if !inputMode {
			formattedInput := formatInput(connectionsInput, cursorPos, inputMode)
			formattedInput = formattedInput + " Max connections"
			cliPrintTb(1, cliTlog.Line, termbox.ColorWhite, termbox.ColorBlack, formattedInput)
		} else {
			formattedInput := formatInput(connectionsInput, cursorPos, inputMode)
			formattedInput = formattedInput + " Max connections"
			cliPrintTb(1, cliTlog.Line, colorCell|termbox.AttrBold, termbox.ColorBlack, formattedInput)
		}
	default:
	}
	cliTlog.Line++
	cliTlog.Line++
	cliTlog.Line++
	cliPrintTb(0, cliTlog.Line, termbox.ColorWhite, termbox.ColorBlack, " Ctrl-Q Quit, Ctrl-S Save, Arrows to navigate, Enter to select, Esc to exit")

	cliTlog.Line = cliTlog.Line + 3
	cliTlog.Print()

	termbox.Flush()

}

func formatInput(input string, cursorPos int, editing bool) string {
	// Formater l'entrée pour qu'elle soit toujours de 6 caractères de large
	formattedInput := fmt.Sprintf("%-6s", input)
	if cursorPos > len(formattedInput) {
		cursorPos = len(formattedInput)
	}
	// Ajouter le curseur à l'endroit approprié si l'utilisateur est en train d'éditer
	if editing {
		return "[" + formattedInput[:cursorPos] + "|" + formattedInput[cursorPos:] + "]"
	} else {
		return "[" + formattedInput + "]"
	}
}
