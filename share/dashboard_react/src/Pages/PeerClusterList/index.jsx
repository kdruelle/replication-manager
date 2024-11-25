import React, { useEffect, useState } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { getClusterPeers } from '../../redux/globalClustersSlice'
import { Box, Flex, HStack, Text } from '@chakra-ui/react'
import NotFound from '../../components/NotFound'
import { AiOutlineCluster } from 'react-icons/ai'
import Card from '../../components/Card'
import TableType2 from '../../components/TableType2'
import styles from './styles.module.scss'
import CheckOrCrossIcon from '../../components/Icons/CheckOrCrossIcon'
import CustomIcon from '../../components/Icons/CustomIcon'
import TagPill from '../../components/TagPill'
import PeerMenu from './PeerMenu'
import { peerLogin, setBaseURL } from '../../redux/authSlice'
import PeerLoginModal from '../../components/Modals/PeerLoginModal'
import { getClusterData, setCluster } from '../../redux/clusterSlice'

function PeerClusterList({ onClick, mode }) {
  const dispatch = useDispatch()

  const [isPeerLoginModalOpen, setIsPeerLoginModalOpen] = useState(false)
  const [clusters, setClusters] = useState([])
  const [url, setURL] = useState("")

  const {
    globalClusters: { loading, clusterPeers },
  } = useSelector((state) => state)

  const openPeerLoginModal = () => {
    setIsPeerLoginModalOpen(true)
  }

  const closePeerLoginModal = () => {
    setIsPeerLoginModalOpen(false)
  }

  const handleEnterCluster = (clusterItem) => {
    dispatch(getClusterData({ clusterName: clusterItem['cluster-name']}))
    if (onClick) {
      onClick(clusterItem)
    }
  }

  const handlePeerLogin = (item) => {
    const token = localStorage.getItem(`user_token_${btoa(item['api-public-url'])}`)
    if (token) {
      dispatch(setBaseURL({ baseURL: item['api-public-url'] }))
      dispatch()
      handleEnterCluster(item)
    } else {
      setURL(item['api-public-url'])
      openPeerLoginModal()
    }
  }

  useEffect(() => {
    dispatch(getClusterPeers({}))
  }, [])

  useEffect(() => {
    if (clusterPeers?.length > 0) {
      if (mode === 'shared') {
        const shared = clusterPeers.filter((cluster) => cluster['cloud18-share'])
        setClusters(shared)
      } else {
        setClusters(clusterPeers)
      }
    }
  }, [clusterPeers])

  return !loading && clusters?.length === 0 ? (
    <NotFound text={mode === 'shared' ? 'No shared peer cluster found!' : 'No peer cluster found!'} />
  ) : (
    <>
      <Flex className={styles.clusterList}>
        {clusters?.map((clusterItem) => {
          const headerText = `${clusterItem['cluster-name']}@${clusterItem['cloud18-domain']}-${clusterItem['cloud18-sub-domain']}-${clusterItem['cloud18-sub-domain-zone']}`

          const dataObject = [
            { key: 'Domain', value: clusterItem['cloud18-domain'] },
            { key: 'Platfom Desciption', value: clusterItem['cloud18-platfom-desciption'] },

            {
              key: 'Share',
              value: (
                <HStack spacing='4'>
                  {clusterItem['cloud18-share'] ? (
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
            }
          ]

          return (
            <Box key={clusterItem['cluster-name']} className={styles.cardWrapper}>
              <Card
                className={styles.card}
                width={'400px'}
                header={
                  <HStack
                    as='button'
                    className={styles.btnHeading}
                    onClick={() => handlePeerLogin(clusterItem)}>
                    <CustomIcon icon={AiOutlineCluster} />
                    <span className={styles.cardHeaderText}>{headerText}</span>

                    <TagPill text='Cloud18' colorScheme='blue' />
                    <PeerMenu colorScheme='blue' clusterItem={clusterItem} className={styles.btnAddUser} labelClassName={styles.rowLabel} valueClassName={styles.rowValue} />
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
      {isPeerLoginModalOpen && <PeerLoginModal baseURL={url} isOpen={isPeerLoginModalOpen} closeModal={closePeerLoginModal} />}
    </>
  )
}

export default PeerClusterList
