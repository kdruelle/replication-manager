import React, { useEffect, useState } from 'react'
import MenuOptions from '../../../components/MenuOptions'
import PeerDetailsModal from '../../../components/Modals/PeerDetailsModal'
import PeerLoginModal from '../../../components/Modals/PeerLoginModal'
import { getClusterData } from '../../../redux/clusterSlice'
import { setBaseURL } from '../../../redux/authSlice'
import { useSelector } from 'react-redux'

function PeerMenu({
    mode,
    onLogin,
    clusterItem,
    className,
    labelClassName,
    valueClassName,
    colorScheme,
}) {
    // const dispatch = useDispatch()
    const [url, setURL] = useState("")
    const [peerDetails, setPeerDetails] = useState(clusterItem)
    const [options, setOptions] = useState([])
    const [isPeerDetailsModalOpen, setIsPeerDetailsModalOpen] = useState(false)
    const [isPeerLoginModalOpen, setIsPeerLoginModalOpen] = useState(false)

    const {
        auth: { loadingPeerLogin, isPeerLogged,baseURL, error }
    } = useSelector((state) => state)


    useEffect(() => {
        if (!loadingPeerLogin) {
            if (isPeerLogged && baseURL != "") {
                handleEnterCluster()
            }
            if (error) {
                setErrorMessage(error)
            }
        }
    }, [loadingPeerLogin])

    const openPeerLoginModal = () => {
        setIsPeerLoginModalOpen(true)
    }

    const closePeerLoginModal = () => {
        setIsPeerLoginModalOpen(false)
    }

    const handleEnterCluster = () => {
        dispatch(getClusterData({ clusterName: clusterItem['cluster-name'] }))
        if (onLogin) {
            onLogin(clusterItem)
        }
    }

    const openPeerDetailsModal = () => {
        setIsPeerDetailsModalOpen(true)
    }
    const closePeerDetailsModal = () => {
        setIsPeerDetailsModalOpen(false)
    }

    const handlePeerLogin = () => {
        const token = localStorage.getItem(`user_token_${btoa(peerDetails['api-public-url'])}`)
        if (token) {
            dispatch(setBaseURL({ baseURL: peerDetails['api-public-url'] }))
            handleEnterCluster(peerDetails)
        } else {
            setURL(peerDetails['api-public-url'])
            openPeerLoginModal()
        }
    }

    useEffect(() => {
        setPeerDetails(clusterItem)
    }, [clusterItem])

    useEffect(() => {
        let opts = []
        if (mode == "shared") {
            opts.push({
                name: 'Register',
                onClick: () => {
                    handlePeerLogin()
                }
            })
        } else {
            opts.push({
                name: 'Login',
                onClick: () => {
                    handlePeerLogin()
                }
            })
        }
        opts.push({
            name: 'Details',
            onClick: () => {
                openPeerDetailsModal()
            }
        })
        setOptions(opts)
    }, [mode, onLogin])

    /* {isConfirmModalOpen && (
        <ConfirmModal
          isOpen={isConfirmModalOpen}
          closeModal={closeConfirmModal}
          title={confirmTitle}
          onConfirmClick={() => {
            confirmHandler()
            closeConfirmModal()
          }}
        />
      )} */

    return (
        <>
            <MenuOptions
                className={className}
                colorScheme={colorScheme}
                placement='left-end'
                options={options}
            />
            {isPeerDetailsModalOpen && (
                <PeerDetailsModal
                    peerDetails={peerDetails}
                    labelClassName={labelClassName}
                    valueClassName={valueClassName}
                    isOpen={isPeerDetailsModalOpen}
                    closeModal={closePeerDetailsModal}
                />
            )}
            {isPeerLoginModalOpen && <PeerLoginModal baseURL={url} isOpen={isPeerLoginModalOpen} closeModal={closePeerLoginModal} />}
        </>
    )
}
export default PeerMenu;