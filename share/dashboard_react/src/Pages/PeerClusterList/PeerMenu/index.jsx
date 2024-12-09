import React, { useEffect, useState } from 'react'
import MenuOptions from '../../../components/MenuOptions'
import PeerDetailsModal from '../../../components/Modals/PeerDetailsModal'
import PeerLoginModal from '../../../components/Modals/PeerLoginModal'
import { getClusterData, peerRegister } from '../../../redux/clusterSlice'
import { peerLogin, setBaseURL } from '../../../redux/authSlice'
import { useDispatch, useSelector } from 'react-redux'

function PeerMenu({
    mode,
    onLogin,
    clusterItem,
    className,
    labelClassName,
    valueClassName,
    colorScheme,
}) {
    const dispatch = useDispatch()
    const [options, setOptions] = useState([])
    const [title, setTitle] = useState("Login to Peer Cluster")
    const [isPeerDetailsModalOpen, setIsPeerDetailsModalOpen] = useState(false)
    const [isPeerLoginModalOpen, setIsPeerLoginModalOpen] = useState(false)

    const {
        auth: { user },
        globalClusters: { monitor }
    } = useSelector((state) => state)

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
        setTitle("Login to Peer Cluster")
        const token = localStorage.getItem(`user_token_${btoa(clusterItem['api-public-url'])}`)
        if (token) {
            dispatch(setBaseURL({ baseURL: clusterItem['api-public-url'] }))
            handleEnterCluster(clusterItem)
        } else {
            openPeerLoginModal()
        }
    }

    const handlePeerRegister = () => {
        setTitle("Register to Peer Cluster")
        openPeerLoginModal()
    }

    const handleSaveModal = (password) => {
        if (mode == "shared") {
            dispatch(peerRegister({ password, clusterName: clusterItem['cluster-name'], baseURL: clusterItem['api-public-url'] }))
        } else {
            dispatch(peerLogin({ password, baseURL: clusterItem['api-public-url'] }))
        }
        closePeerLoginModal()
    }

    useEffect(() => {
        let opts = []
        // For safety, user should use email if they want to login to peer
        if ("admin" != user?.username) {
            if (mode == "shared") {
                opts.push({
                    name: 'Register',
                    onClick: () => {
                        handlePeerRegister()
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
        }
        setOptions(opts)
    }, [mode, onLogin])

    if (user?.username == "admin") {
        return (<></>)
    } else {
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
                        peerDetails={clusterItem}
                        labelClassName={labelClassName}
                        valueClassName={valueClassName}
                        isOpen={isPeerDetailsModalOpen}
                        closeModal={closePeerDetailsModal}
                    />
                )}
                {isPeerLoginModalOpen && <PeerLoginModal title={title} isOpen={isPeerLoginModalOpen} closeModal={closePeerLoginModal} onSaveModal={handleSaveModal} />}
            </>
        )
    }
}
export default PeerMenu;