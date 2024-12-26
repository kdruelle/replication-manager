import React, { useEffect, useState } from 'react'
import MenuOptions from '../../../components/MenuOptions'
import PeerSubscribeModal from '../../../components/Modals/PeerSubscribeModal'
import { getClusterData, peerRegister } from '../../../redux/clusterSlice'
import { peerLogin, setBaseURL } from '../../../redux/authSlice'
import { useDispatch, useSelector } from 'react-redux'

function PeerMenu({
    mode,
    onLogin,
    clusterItem,
    className,
    colorScheme,
}) {
    const dispatch = useDispatch()
    const [options, setOptions] = useState([])
    const [title, setTitle] = useState("Login to Peer Cluster")
    const [isPeerSubscribeModalOpen, setIsPeerSubscribeModalOpen] = useState(false)

    const {
        auth: { user },
        globalClusters: { monitor }
    } = useSelector((state) => state)

    const openPeerSubscribeModal = () => {
        setIsPeerSubscribeModalOpen(true)
    }

    const closePeerSubscribeModal = () => {
        setIsPeerSubscribeModalOpen(false)
    }

    const handleEnterCluster = () => {
        dispatch(getClusterData({ clusterName: clusterItem['cluster-name'] }))
        if (onLogin) {
            onLogin(clusterItem)
        }
    }

    const handlePeerLogin = () => {
        setTitle("Login to Peer Cluster")
        const token = localStorage.getItem(`user_token_${btoa(clusterItem['api-public-url'])}`)
        if (token) {
            dispatch(setBaseURL({ baseURL: clusterItem['api-public-url'] }))
            handleEnterCluster(clusterItem)
        } else {
            openPeerSubscribeModal()
        }
    }

    const handlePeerRegister = () => {
        setTitle("Register to Peer Cluster")
        openPeerSubscribeModal()
    }

    const handleSaveModal = (password) => {
        if (mode == "shared") {
            dispatch(peerRegister({ password, clusterName: clusterItem['cluster-name'], baseURL: clusterItem['api-public-url'] }))
        } else {
            dispatch(peerLogin({ password, baseURL: clusterItem['api-public-url'] }))
        }
        closePeerSubscribeModal()
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
                {isPeerSubscribeModalOpen && <PeerSubscribeModal title={title} isOpen={isPeerSubscribeModalOpen} closeModal={closePeerSubscribeModal} onSaveModal={handleSaveModal} />}
            </>
        )
    }
}
export default PeerMenu;