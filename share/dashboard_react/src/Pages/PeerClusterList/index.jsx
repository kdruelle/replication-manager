import React, { useEffect, useState } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { getClusterPeers, getClusterForSale, getTermsData } from '../../redux/globalClustersSlice'
import { Box, Flex, HStack, Text, Wrap } from '@chakra-ui/react'
import NotFound from '../../components/NotFound'
import { AiOutlineCluster } from 'react-icons/ai'
import Card from '../../components/Card'
import TableType2 from '../../components/TableType2'
import styles from './styles.module.scss'
import CheckOrCrossIcon from '../../components/Icons/CheckOrCrossIcon'
import CustomIcon from '../../components/Icons/CustomIcon'
import TagPill from '../../components/TagPill'
import { HiCreditCard, HiTag } from 'react-icons/hi'
import { peerLogin, setBaseURL } from '../../redux/authSlice'
import { getClusterData, peerSubscribe } from '../../redux/clusterSlice'
import PeerSubscribeModal from '../../components/Modals/PeerSubscribeModal'
import { showErrorToast } from '../../redux/toastSlice'

function PeerClusterList({ onLogin, mode }) {
  const dispatch = useDispatch()
  const [clusters, setClusters] = useState([])
  const [item, setItem] = useState({})
  const [isPeerSubscribeModalOpen, setIsPeerSubscribeModalOpen] = useState(false)

  const {
    globalClusters: { loading, clusterPeers, clusterForSale, monitor, terms },
    auth: {
      user
    },
  } = useSelector((state) => state)

  useEffect(() => {
    dispatch(getClusterPeers({}))
    dispatch(getClusterForSale({}))
  }, [])

  useEffect(() => {
    dispatch(getTermsData({}))
  }, [monitor?.termsDT])

  useEffect(() => {
    if (clusterPeers?.length > 0 && mode !== 'shared') {
      setClusters(clusterPeers)
    }
    if (clusterForSale?.length > 0 && mode === 'shared') {
      setClusters(clusterForSale)
    }
  }, [clusterPeers,clusterForSale])

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
    dispatch(peerSubscribe({ clusterName: clusterItem['cluster-name'], baseURL: clusterItem['api-public-url'] }))
  }

  const handlePeerCluster = (clusterItem, isRelogin = false) => {
    let handler
    let baseURL = clusterItem['api-public-url']
    let token = localStorage.getItem(`user_token`)

    if (monitor?.config?.apiPublicUrl == baseURL) {
      baseURL = ''
    }

    if (baseURL !== '') {
      token = localStorage.getItem(`user_token_${btoa(baseURL)}`)
    }

    if (token && !isRelogin) {
      dispatch(setBaseURL({ baseURL: baseURL }));
      handler = dispatch(getClusterData({ clusterName: clusterItem['cluster-name'] }))
    } else {
      localStorage.removeItem(`user_token_${btoa(baseURL)}`)
      handler = dispatch(peerLogin({ baseURL: baseURL }))
        .then((action) => {
          if (action?.payload?.status === 200) {
            return dispatch(getClusterData({ clusterName: clusterItem['cluster-name'] }))
          } else {
            dispatch(
              showErrorToast({
                status: 'error',
                title: 'Peer login failed',
                description: action?.payload?.data || error
              })
            )
            dispatch(setBaseURL({ baseURL: '' }));
            throw new Error(action?.payload?.data);
          }
        });
    }

    handler.then((resp) => {
      // Handle peer relogin if peer repman instance was restarted
      if (!isRelogin && resp?.payload?.status === 401 && resp?.payload?.data.includes("crypto/rsa: verification error")) {
        return handlePeerCluster(clusterItem, true)
      }

      if (mode === "shared") {
        setItem(clusterItem);
        openPeerSubscribeModal();
      } else {
        if (resp?.payload?.status === 200) {
          if (onLogin) return onLogin(resp.payload.data);
        }

        dispatch(setBaseURL({ baseURL: '' }));
        showErrorToast({
          status: 'error',
          title: 'Peer login failed',
          description: resp?.payload?.data || "Peer login failed"
        })
      }
    })
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

          const isPending = clusterItem?.['api-credentials-acl-allow']?.includes('pending')
        const isSponsor = clusterItem?.['api-credentials-acl-allow']?.includes('sponsor')

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
            { key: 'Certifications', value: clusterItem['cloud18-infra-certifications']  },
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
                    <CustomIcon icon={ isSponsor || isPending ? (HiCreditCard): (AiOutlineCluster)} fill={ isSponsor ? "green" : isPending ? "orange" : "gray" }  />
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
      {isPeerSubscribeModalOpen && <PeerSubscribeModal terms={terms} cluster={item} user={user} isOpen={isPeerSubscribeModalOpen} closeModal={closePeerSubscribeModal} onSaveModal={handleSubscribeModal} />}
    </>
  )
}

export default PeerClusterList
