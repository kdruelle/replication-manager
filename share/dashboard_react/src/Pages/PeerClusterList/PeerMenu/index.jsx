import React, { useEffect, useState } from 'react'
import MenuOptions from '../../../components/MenuOptions'
import PeerDetailsModal from '../../../components/Modals/PeerDetailsModal'

function PeerMenu({
    clusterItem,
    className,
    labelClassName,
    valueClassName,
    colorScheme,
}) {
    // const dispatch = useDispatch()
    const [peerDetails, setPeerDetails] = useState(clusterItem)
    const [isPeerDetailsModalOpen, setIsPeerDetailsModalOpen] = useState(false)

    // const openConfirmModal = () => {
    //   setIsConfirmModalOpen(true)
    // }
    // const closeConfirmModal = () => {
    //   setIsConfirmModalOpen(false)
    //   setConfirmHandler(null)
    //   setConfirmTitle('')
    // }

    const openPeerDetailsModal = () => {
        setIsPeerDetailsModalOpen(true)
    }
    const closePeerDetailsModal = () => {
        setIsPeerDetailsModalOpen(false)
    }

    useEffect(() => {
        setPeerDetails(clusterItem)
    }, [clusterItem])

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
                options={[
                    {
                        name: 'Details',
                        onClick: () => {
                            openPeerDetailsModal()
                        }
                    },
                ]}
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
        </>
    )
}
export default PeerMenu;