import React, { useEffect, useState } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { getClusterPeers, getTermsData } from '../../redux/globalClustersSlice'
import { Box, Flex, HStack, Text, Wrap } from '@chakra-ui/react'
import NotFound from '../../components/NotFound'
import { AiOutlineCluster } from 'react-icons/ai'
import Card from '../../components/Card'
import TableType2 from '../../components/TableType2'
import styles from './styles.module.scss'
import CheckOrCrossIcon from '../../components/Icons/CheckOrCrossIcon'
import CustomIcon from '../../components/Icons/CustomIcon'
import TagPill from '../../components/TagPill'
import { HiTag } from 'react-icons/hi'
import { peerLogin, setBaseURL } from '../../redux/authSlice'
import { getClusterData, peerRegister } from '../../redux/clusterSlice'
import PeerSubscribeModal from '../../components/Modals/PeerSubscribeModal'
import { showErrorToast } from '../../redux/toastSlice'

function PeerClusterList({ onLogin, mode }) {
  const dispatch = useDispatch()
  const [clusters, setClusters] = useState([])
  const [item, setItem] = useState({})
  const [isPeerSubscribeModalOpen, setIsPeerSubscribeModalOpen] = useState(false)

  const {
    globalClusters: { loading, clusterPeers, monitor, terms },
    auth: {
      user
    },
  } = useSelector((state) => state)

  useEffect(() => {
    dispatch(getClusterPeers({}))
  }, [])

  useEffect(() => {
      dispatch(getTermsData({}))
  }, [monitor?.termsDT])

  // Regular expression to check if the string is an email
  const emailPattern = /^[a-zA-Z0-9._-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;

  const checkPeerUser = (u, gituser) => {


    // Check if the string matches the email pattern or is "admin"
    if (emailPattern.test(u)) {
      return u;
    } else if (u.toLowerCase() === "admin") {
      return gituser || "";
    }
    return "";
  }

  useEffect(() => {
    if (clusterPeers?.length > 0) {
      if (mode === 'shared') {
        const shared = clusterPeers.filter((cluster) => cluster["cloud18-shared"] === "true" && cluster["cloud18-peer"] === "false")
        setClusters(shared)
      } else {
        const peers = user?.username ? clusterPeers.filter((cluster) => cluster["cloud18-peer"] === "true").filter((cluster) => cluster['api-credentials-acl-allow']?.includes(checkPeerUser(user?.username, monitor?.config?.cloud18GitUser))) : []
        setClusters(peers)
      }
    }
  }, [clusterPeers])

  const openPeerSubscribeModal = () => {
    setIsPeerSubscribeModalOpen(true)
  }

  const closePeerSubscribeModal = (keepBaseURL = false) => {
    if (!keepBaseURL) {
      dispatch(setBaseURL({ baseURL: '' }))
    }
    setIsPeerSubscribeModalOpen(false)
  }

  const handleSubscribeModal = (clusterItem) => {
    closePeerSubscribeModal(true)
    dispatch(peerRegister({ clusterName: clusterItem['cluster-name'], baseURL: clusterItem['api-public-url'] }))
  }

  const handlePeerCluster = (clusterItem) => {
    const token = localStorage.getItem(`user_token_${btoa(clusterItem['api-public-url'])}`)
    let handler
    if (token) {
      dispatch(setBaseURL({ baseURL: clusterItem['api-public-url'] }));
      handler = dispatch(getClusterData({ clusterName: clusterItem['cluster-name'] }))
    } else {
      handler = dispatch(peerLogin({ baseURL: clusterItem['api-public-url'] }))
        .then((action) => {
          if (action?.payload?.status === 200) {
            return dispatch(getClusterData({ clusterName: clusterItem['cluster-name'] }))
          } else {
            throw new Error(action?.payload?.data);
          }
        });
    }

    handler.then((resp) => {
      if (resp?.payload?.status === 200) {
        if (onLogin) return onLogin(resp.payload.data);
      }

      if (mode === "shared") {
        setItem(clusterItem);
        openPeerSubscribeModal();
      } else {
        dispatch(setBaseURL({ baseURL: '' }));
        showErrorToast({
          status: 'error',
          title: 'Peer login failed',
          description: resp?.payload?.data || "Peer login failed"
        })
      }
    })
      .catch((error) => {
        dispatch(
          showErrorToast({
            status: 'error',
            title: 'Peer login failed',
            description: error?.message || error
          })
        )
        dispatch(setBaseURL({ baseURL: '' }));
      });
  };


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
          const cost = clusterItem['cloud18-monthly-infra-cost'] * 1 + clusterItem['cloud18-monthly-license-cost'] * 1 + clusterItem['cloud18-monthly-sysops-cost'] * 1 + clusterItem['cloud18-monthly-dbops-cost'] * 1
          const amount = (cost * (100 - clusterItem['cloud18-promotion-pct'])) / 100
          const currency = clusterItem['cloud18-cost-currency']

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
            {
              key: (
                <HStack spacing='4'>
                  {clusterItem['cloud18-promotion-pct'] && clusterItem['cloud18-promotion-pct'] > 0 ? (
                    <>
                      <Text>Price</Text>
                      <CustomIcon color={"red"} icon={HiTag} />
                    </>
                  ) : (
                    <>
                      <Text>Price</Text>
                    </>
                  )}
                </HStack>
              ), value: (
                <HStack spacing='4'>
                  {clusterItem['cloud18-promotion-pct'] && clusterItem['cloud18-promotion-pct'] > 0 ? (
                    <>
                      <Text>
                        <Text as={"span"} textColor="red.500" textDecorationColor="red.500" textDecoration="line-through">
                          {cost.toFixed(2)}
                        </Text>
                        &nbsp;
                        <Text as={"span"} fontWeight="bold">
                          {amount.toFixed(2)} {currency}/Month
                        </Text>
                      </Text>
                    </>
                  ) : (
                    <>
                      <Text>{cost.toFixed(2)} {currency}/Month</Text>
                    </>
                  )}
                </HStack>
              )
            },
            { key: 'Memory', value: clusterItem['prov-db-memory'] / 1024 + "GB" },
            { key: 'IOps', value: clusterItem['prov-db-disk-iops'] },
            { key: 'Disk', value: clusterItem['prov-db-disk-size'] + "GB" },
            { key: 'CPU Core', value: clusterItem['prov-db-cpu-cores'] },
            { key: 'CPU Type', value: clusterItem['cloud18-infra-cpu-model'] },
            { key: 'CPU Freq', value: clusterItem['cloud18-infra-cpu-freq'] },
            { key: 'Data Centers', value: clusterItem['cloud18-infra-data-centers'] },
            { key: 'Public Bandwidth', value: clusterItem['cloud18-infra-public-bandwidth'] / 1024 + "Gbps" },
            { key: 'Time To Response', value: clusterItem['cloud18-sla-response-time'] + "Hours" },
            { key: 'Time To Repair', value: clusterItem['cloud18-sla-repair-time'] + "Hours" },
            { key: 'Time To Provision', value: clusterItem['cloud18-sla-provision-time'] + "Hours" },
            { key: 'Infrastructure', value: clusterItem['prov-orchestrator'] + " " + clusterItem['cloud18-platform-description'] },
            /*  {
                key: 'Share',
                value: (
                  <HStack spacing='4'>
                    {clusterItem['cloud18-is-multi-dc'] ? (
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
                    as="button"
                    className={styles.btnHeading}
                    onClick={() => { handlePeerCluster(clusterItem) }}>
                    <CustomIcon icon={AiOutlineCluster} />
                    <span className={styles.cardHeaderText}>{headerText}</span>
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
      {isPeerSubscribeModalOpen && <PeerSubscribeModal terms={terms} cluster={item} isOpen={isPeerSubscribeModalOpen} closeModal={closePeerSubscribeModal} onSaveModal={handleSubscribeModal} />}
    </>
  )
}

export default PeerClusterList
