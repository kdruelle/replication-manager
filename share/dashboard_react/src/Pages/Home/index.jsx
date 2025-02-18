import React, { useEffect, useRef, useState } from 'react'
import { Box } from '@chakra-ui/react'

import PageContainer from '../PageContainer'
import TabItems from '../../components/TabItems'
import ClusterList from '../ClusterList'
import { useDispatch, useSelector } from 'react-redux'
import {
  getBackupSnapshot,
  getClusterAlerts,
  getClusterCertificates,
  getClusterData,
  getClusterMaster,
  getClusterProxies,
  getClusterServers,
  getJobs,
  getQueryRules,
  getShardSchema,
  getTopProcess,
  setCluster,
  setRefreshInterval,
  pauseAutoReload
} from '../../redux/clusterSlice'
import { getClusters, getMonitoredData, getClusterPeers } from '../../redux/globalClustersSlice'
import { AppSettings } from '../../AppSettings'
import styles from './styles.module.scss'
import { useParams } from 'react-router-dom'
import { HiArrowNarrowLeft } from 'react-icons/hi'
import CustomIcon from '../../components/Icons/CustomIcon'
import Dashboard from '../Dashboard'
import Settings from '../Settings'
import Configs from '../Configs'
import Graphs from '../Graphs'
import Agents from '../Agents'
import Users from '../Users'
import Maintenance from '../Maintenance'
import Top from '../Top'
import Shards from '../Shards'
import QueryRules from '../QueryRules'
import PeerClusterList from '../PeerClusterList'
import ClustersGlobalSettings from '../ClustersGlobalSettings'
import NewClusterModal from '../../components/Modals/NewClusterModal'
import { FaPlus } from 'react-icons/fa'
import RMIconButton from '../../components/RMIconButton'
import { setBaseURL } from '../../redux/authSlice'

function Home() {
  const dispatch = useDispatch()
  const selectedTabRef = useRef(0)
  const selectedClusterNameRef = useRef('')
  const isClusterOpenRef = useRef(false)
  const [selectedTab, setSelectedTab] = useState(0)
  const [user, setUser] = useState(null)
  const [selectedCluster, setSelectedCluster] = useState(null)
  const dashboardTabsRef = useRef([])
  const globalTabsRef = useRef([])
  const [isNewClusterModalOpen, setIsNewClusterModalOpen] = useState(false)

  const params = useParams()

  const {
    cluster: { refreshInterval, clusterData },
    globalClusters: { monitor }
  } = useSelector((state) => state)

  useEffect(() => {
    if (params?.cluster) {
      setDashboardTab({ name: params.cluster })
    }
  }, [])

  useEffect(() => {
    const loggedUser = localStorage.getItem('username')
    if (monitor?.config?.cloud18) {
      globalTabsRef.current = ['Clusters Local', 'Clusters Peer', 'Clusters For Sale']
    } else {
      globalTabsRef.current = ['Clusters Local']
    }
    if (loggedUser == "admin") {
      globalTabsRef.current.push('Settings')
    }
  }, [monitor?.config?.cloud18])

  useEffect(() => {
    if (clusterData) {
      setSelectedCluster(clusterData)
      if (clusterData.apiUsers) {
        const loggedUser = localStorage.getItem('username')
        if (loggedUser && clusterData?.apiUsers[loggedUser]) {
          const apiUser = clusterData.apiUsers[loggedUser]
          setUser(apiUser)
          const authorizedTabs = ['Dashboard', 'Settings', 'Configs']
          if (clusterData.config.graphiteMetrics && apiUser.grants['cluster-show-graphs']) {
            authorizedTabs.push('Graphs')
          }
          if (apiUser.grants['cluster-show-agents']) {
            authorizedTabs.push('Agents')
          }
          if (apiUser.grants['cluster-show-backups']) {
            authorizedTabs.push('Maintenance')
          }
          if (apiUser.grants['db-show-process']) {
            authorizedTabs.push('Tops')
          }
          if (clusterData.config.proxysql && apiUser.grants['cluster-show-agents']) {
            authorizedTabs.push('Query Rules')
          }
          if (apiUser.grants['db-show-schema']) {
            authorizedTabs.push('Shards')
          }
          if (apiUser.grants['cluster-grant']) {
            authorizedTabs.push('Users')
          }
          dashboardTabsRef.current = authorizedTabs
        }
      }
    }
  }, [clusterData])

  useEffect(() => {
    let intervalId = 0
    let interval = localStorage.getItem('refresh_interval')
      ? parseInt(localStorage.getItem('refresh_interval'))
      : AppSettings.DEFAULT_INTERVAL

    dispatch(setRefreshInterval({ interval }))

    if (refreshInterval > 0) {
      callServices()
      const intervalSeconds = refreshInterval * 1000
      intervalId = setInterval(() => {
        callServices()
      }, intervalSeconds)
    }

    return () => {
      clearInterval(intervalId)
    }
  }, [refreshInterval])

  const renderClusterListTabWithArrow = () => {
    return (
      <>
        <CustomIcon icon={HiArrowNarrowLeft} /> Clusters
      </>
    )
  }

  const callServices = () => {
    const isAutoReloadPaused = localStorage.getItem('pause_auto_reload')

    if (!isClusterOpenRef.current) {
      if (
        globalTabsRef.current[selectedTabRef.current] === 'Clusters Local' ||
        globalTabsRef.current[selectedTabRef.current] === 'Settings'
      ) {
        if (!isAutoReloadPaused) {
          dispatch(getMonitoredData({}))
          dispatch(getClusters({}))
        }
      }
      if (
        globalTabsRef.current[selectedTabRef.current] === 'Clusters Peer' ||
        globalTabsRef.current[selectedTabRef.current] === 'Clusters For Sale'
      ) {
        dispatch(getClusterPeers({}))
      }
    } else if (selectedClusterNameRef.current) {
      if (!isAutoReloadPaused) {
        dispatch(getClusterData({ clusterName: selectedClusterNameRef.current }))
        dispatch(getClusterAlerts({ clusterName: selectedClusterNameRef.current }))
        dispatch(getClusterMaster({ clusterName: selectedClusterNameRef.current }))
        dispatch(getClusterServers({ clusterName: selectedClusterNameRef.current }))
        dispatch(getClusterProxies({ clusterName: selectedClusterNameRef.current }))
      }
      if (dashboardTabsRef.current[selectedTabRef.current - 1] === 'Configs') {
        dispatch(getClusterCertificates({ clusterName: selectedClusterNameRef.current }))
      }
      if (dashboardTabsRef.current[selectedTabRef.current - 1] === 'Maintenance') {
        dispatch(getBackupSnapshot({ clusterName: selectedClusterNameRef.current }))
        dispatch(getJobs({ clusterName: selectedClusterNameRef.current }))
      }
      if (dashboardTabsRef.current[selectedTabRef.current - 1] === 'Tops') {
        dispatch(getTopProcess({ clusterName: selectedClusterNameRef.current }))
      }
      if (dashboardTabsRef.current[selectedTabRef.current - 1] === 'Query Rules') {
        dispatch(getQueryRules({ clusterName: selectedClusterNameRef.current }))
      }
      if (dashboardTabsRef.current[selectedTabRef.current - 1] === 'Shards') {
        dispatch(getShardSchema({ clusterName: selectedClusterNameRef.current }))
      }
    }
  }
  const handleTabChange = (tabIndex) => {
    selectedTabRef.current = tabIndex
    setSelectedTab(tabIndex)
    if (tabIndex === 0) {
      isClusterOpenRef.current = false
      dispatch(setCluster({ data: null }))
      dispatch(setBaseURL({ baseURL: '' }))
      selectedClusterNameRef.current = ''
    }
  }

  const setDashboardTab = (cluster) => {
    selectedTabRef.current = 1
    isClusterOpenRef.current = true
    selectedClusterNameRef.current = cluster.name
    setSelectedTab(1)
  }
  const openNewClusterModal = (e) => {
    e.stopPropagation()
    setIsNewClusterModalOpen(true)
    setSelectedTab(0)
    dispatch(pauseAutoReload({ isPaused: true }))
  }

  const closeNewClusterModal = () => {
    setIsNewClusterModalOpen(false)
    dispatch(pauseAutoReload({ isPaused: false }))
  }

  return (
    <PageContainer>
      <Box className={styles.container}>
        <TabItems
          tabIndex={selectedTab}
          onChange={handleTabChange}
          options={
            isClusterOpenRef.current
              ? [renderClusterListTabWithArrow(), ...dashboardTabsRef.current]
              : globalTabsRef.current
          }
          tabPrefix={selectedClusterNameRef.current == '' && (<div onClick={openNewClusterModal} className={styles.tabSelected}><CustomIcon icon={FaPlus}/></div>)}
          tabContents={[
            <ClusterList onClick={setDashboardTab} />,
            ...(isClusterOpenRef.current
              ? [
                <Dashboard user={user} selectedCluster={selectedCluster} />,
                <Settings user={user} selectedCluster={selectedCluster} />,
                <Configs user={user} selectedCluster={selectedCluster} />,
                ...(selectedCluster?.config?.graphiteMetrics && user?.grants['cluster-show-graphs']
                  ? [<Graphs />]
                  : []),
                ...(user?.grants['cluster-show-agents']
                  ? [<Agents user={user} selectedCluster={selectedCluster} />]
                  : []),
                ...(user?.grants['cluster-show-backups']
                  ? [<Maintenance user={user} selectedCluster={selectedCluster} />]
                  : []),
                ...(user?.grants['db-show-process'] ? [<Top selectedCluster={selectedCluster} />] : []),
                ...(selectedCluster?.config?.proxysql && user?.grants['cluster-show-agents']
                  ? [<QueryRules selectedCluster={selectedCluster} />]
                  : []),
                ...(user?.grants['db-show-schema'] ? [<Shards selectedCluster={selectedCluster} />] : []),
                ...(user?.grants['cluster-grant'] ? [<Users selectedCluster={selectedCluster} user={user}/>] : [])
              ]
              : globalTabsRef.current.includes('Clusters Peer') // monitor?.config?.cloud18 is false, do not show "Peer Clusters" tab
                ? [<PeerClusterList onLogin={setDashboardTab} />, <PeerClusterList onLogin={setDashboardTab} mode='shared' />, <ClustersGlobalSettings />]
                : [<ClustersGlobalSettings />])
          ]}
        />
        {
          selectedClusterNameRef.current == '' && (
            <>
              {isNewClusterModalOpen && (
                <NewClusterModal plans={monitor?.servicePlans} orchestrators={monitor?.serviceOrchestrators} defaultOrchestrator={monitor?.config.provOrchestrator} isOpen={isNewClusterModalOpen} closeModal={closeNewClusterModal} />
              )}
            </>
          )
        }
      </Box>
    </PageContainer>
  )
}

export default Home
