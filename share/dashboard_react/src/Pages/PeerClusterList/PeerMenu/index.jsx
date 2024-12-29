import React, { useEffect, useState } from 'react'
import MenuOptions from '../../../components/MenuOptions'
import ClusterSubscribeModal from '../../../components/Modals/ClusterSubscribeModal'
import { getClusterData, clusterSubscribe } from '../../../redux/clusterSlice'
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
    const [isClusterSubscribeModalOpen, setIsClusterSubscribeModalOpen] = useState(false)

    const {
        auth: { user },
        globalClusters: { monitor }
    } = useSelector((state) => state)

    const openClusterSubscribeModal = () => {
        setIsClusterSubscribeModalOpen(true)
    }

    const closeClusterSubscribeModal = () => {
        setIsClusterSubscribeModalOpen(false)
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
            openClusterSubscribeModal()
        }
    }

    const handleclusterSubscribe = () => {
        setTitle("Register to Peer Cluster")
        openClusterSubscribeModal()
    }

    const handleSaveModal = (password) => {
        if (mode == "shared") {
            dispatch(clusterSubscribe({ password, clusterName: clusterItem['cluster-name'], baseURL: clusterItem['api-public-url'] }))
        } else {
            dispatch(peerLogin({ password, baseURL: clusterItem['api-public-url'] }))
        }
        closeClusterSubscribeModal()
    }

    useEffect(() => {
        let opts = []
        // For safety, user should use email if they want to login to peer
        if ("admin" != user?.username) {
            if (mode == "shared") {
                opts.push({
                    name: 'Register',
                    onClick: () => {
                        handleclusterSubscribe()
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
                {isClusterSubscribeModalOpen && <ClusterSubscribeModal title={title} isOpen={isClusterSubscribeModalOpen} closeModal={closeClusterSubscribeModal} onSaveModal={handleSaveModal} />}
            </>
        )
    }
}
export default PeerMenu;