import React, { useEffect, useState } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { getClusterPeers } from '../../redux/globalClustersSlice'
import { Box, Flex, HStack, Text, Wrap } from '@chakra-ui/react'
import NotFound from '../../components/NotFound'
import { AiOutlineCluster } from 'react-icons/ai'
import Card from '../../components/Card'
import TableType2 from '../../components/TableType2'
import styles from './styles.module.scss'
import CheckOrCrossIcon from '../../components/Icons/CheckOrCrossIcon'
import CustomIcon from '../../components/Icons/CustomIcon'
import TagPill from '../../components/TagPill'
import PeerMenu from './PeerMenu'

function PeerClusterList({ onLogin, mode }) {
  const dispatch = useDispatch()
  const [clusters, setClusters] = useState([])

  const {
    globalClusters: { loading, clusterPeers, monitor },
    auth: {
      user
    },
  } = useSelector((state) => state)

  useEffect(() => {
    dispatch(getClusterPeers({}))
  }, [])

  const checkPeerACL = (u, gituser) => {
    // Regular expression to check if the string is an email
    const emailPattern = /^[a-zA-Z0-9._-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;

    // Check if the string matches the email pattern or is "admin"
    if (emailPattern.test(u)) {
      return u;
    } else if (u.toLowerCase() === "admin") {
      return gituser;
    }
    return "";
  }

  useEffect(() => {
    if (clusterPeers?.length > 0) {
      if (mode === 'shared') {
        const shared = clusterPeers.filter((cluster) => cluster["cloud18-shared"] === "true" && cluster["cloud18-peer"] === "false")
        setClusters(shared)
      } else {
        const peers = user?.username ? clusterPeers.filter((cluster) => cluster["cloud18-peer"] === "true").filter((cluster) => cluster['api-credentials-acl-allow']?.includes(checkPeerACL(user?.username, monitor?.config?.cloud18GitUser || ""))) : []
        setClusters(peers)
      }
    }
  }, [clusterPeers])

  return !loading && clusters?.length === 0 ? (
    <NotFound text={mode === 'shared' ? 'No shared peer cluster found!' : 'No peer cluster found!'} />
  ) : (
    <>
      <Flex className={styles.clusterList}>
        {clusters?.map((clusterItem) => {
          const headerText = `${clusterItem['cluster-name']}\n`
          const domain = `${clusterItem['cloud18-domain']}`
          const subDomain = `${clusterItem['cloud18-sub-domain']}`
          const subDomainZone = ` ${clusterItem['cloud18-sub-domain-zone']}`
          const cost = clusterItem['cloud18-monthly-infra-cost']*1+clusterItem['cloud18-monthly-license-cost']*1 + clusterItem['cloud18-monthly-sysops-cost']*1 + clusterItem['cloud18-monthly-dbops-cost']*1
          const currency  = clusterItem['cloud18-cost-currency']
          const price = `${cost} ${currency}/Month`

          const dataObject = [
            {
              key: 'Tags', value: (
                <>
                  <Wrap>
                    <TagPill text='cloud18' colorScheme='blue' />
                    <TagPill text={domain} colorScheme='blue' />
                    <TagPill text={subDomain} colorScheme='blue' />
                    <TagPill text={subDomainZone} colorScheme='blue' />
                  </Wrap>
                </>
              )
            },
            { key: 'Service Plan', value: clusterItem['prov-service-plan'] },
            { key: 'Geo Zone', value: clusterItem['cloud18-infra-geo-localizations'] },
            { key: 'Memory', value: clusterItem['prov-db-memory']/1024 + "GB"},
            { key: 'IOps', value: clusterItem['prov-db-disk-iops'] },
            { key: 'Disk', value: clusterItem['prov-db-disk-size']+ "GB" },
            { key: 'CPU Core', value: clusterItem['prov-db-cpu-cores']},
            { key: 'CPU Type', value: clusterItem['cloud18-infra-cpu-model'] },
            { key: 'Data Centers', value: clusterItem['cloud18-infra-data-centers'] },
            { key: 'Public Bandwidth', value: clusterItem['cloud18-infra-public-bandwidth']/1024 +"Gbps"},
            { key: 'Price', value: price  },
            { key: 'Infrastructure', value: clusterItem['prov-orchestrator']  + " " + clusterItem['cloud18-platform-description'] },

            /*  {
                key: 'Share',
                value: (
                  <HStack spacing='4'>
                    {clusterItem['cloud18-shared'] ? (
                      <>
                        <CheckOrCrossIcon isValid={true} />
                        <Text>Yes</Text>
                      </>
                    ) : (
                      <>
                        <CheckOrCrossIcon isValid={false} />
                        <Text>No</Text>
                      </>
                    )}

                  </HStack>
                )
              }*/
          ]

          return (
            <Box key={clusterItem['cluster-name']} className={styles.cardWrapper}>
              <Card
                className={styles.card}
                width={'400px'}
                header={
                  <HStack
                    as='button'
                    className={styles.btnHeading}>
                    <CustomIcon icon={AiOutlineCluster} />
                    <span className={styles.cardHeaderText}>{headerText}</span>

                    <PeerMenu mode={mode} onLogin={onLogin} colorScheme='blue' clusterItem={clusterItem} className={styles.btnAddUser} labelClassName={styles.rowLabel} valueClassName={styles.rowValue} />

                  </HStack>
                }
                body={

                  <TableType2
                    dataArray={dataObject}
                    className={styles.table}
                    labelClassName={styles.rowLabel}
                    valueClassName={styles.rowValue}
                  />
                }
              />
            </Box>
          )
        })}
      </Flex>
    </>
  )
}

export default PeerClusterList
