import {
    Modal,
    ModalBody,
    ModalCloseButton,
    ModalContent,
    ModalHeader,
    ModalOverlay,
    theme,
} from '@chakra-ui/react'
import React, { useEffect, useMemo } from 'react'
import parentStyles from './styles.module.scss'
import TableType2 from '../TableType2'

function PeerDetailsModal({ peerDetails, tableClassName, labelClassName, valueClassName, isOpen, closeModal }) {

    useEffect(() => {},[peerDetails])
    const dataObject = [
        { key: "cloud18-domain", value: peerDetails ? peerDetails["cloud18-domain"] : "" },
        { key: "cloud18-sub-domain", value: peerDetails ? peerDetails["cloud18-sub-domain"] : "" },
        { key: "cloud18-sub-domain-zone", value: peerDetails ? peerDetails["cloud18-sub-domain-zone"] : "" },
        { key: "cloud18-plan", value: peerDetails ? peerDetails["cloud18-plan"] : "" },
        { key: "cloud18-montly-infra-cost", value: peerDetails ? peerDetails["cloud18-montly-infra-cost"] : "" },
        { key: "cloud18-montly-license-cost", value: peerDetails ? peerDetails["cloud18-montly-license-cost"] : "" },
        { key: "cloud18-montly-sysops-cost", value: peerDetails ? peerDetails["cloud18-montly-sysops-cost"] : "" },
        { key: "cloud18-montly-dbops-cost", value: peerDetails ? peerDetails["cloud18-montly-dbops-cost"] : "" },
        { key: "cloud18-cost-currency", value: peerDetails ? peerDetails["cloud18-cost-currency"] : "" },
        { key: "cloud18-open-dbops", value: peerDetails ? peerDetails["cloud18-open-dbops"] : "" },
        { key: "cloud18-subscribed-dbops", value: peerDetails ? peerDetails["cloud18-subscribed-dbops"] : "" },
        { key: "cloud18-open-sysops", value: peerDetails ? peerDetails["cloud18-open-sysops"] : "" },
        { key: "cloud18-database-write-srv-record", value: peerDetails ? peerDetails["cloud18-database-write-srv-record"] : "" },
        { key: "cloud18-database-read-srv-record", value: peerDetails ? peerDetails["cloud18-database-read-srv-record"] : "" },
        { key: "cloud18-database-read-write-srv-record", value: peerDetails ? peerDetails["cloud18-database-read-write-srv-record"] : "" },
    ]

    return (
        <Modal size={'xl'} isOpen={isOpen} onClose={closeModal}>
            <ModalOverlay />
            <ModalContent className={parentStyles.modalLightContent}>
                <ModalHeader>{'Peer Details'}</ModalHeader>
                <ModalCloseButton />
                <ModalBody>
                    <TableType2 dataArray={dataObject} className={parentStyles.table} labelClassName={labelClassName} valueClassName={valueClassName} />
                </ModalBody>
            </ModalContent>
        </Modal>
    )
}

export default PeerDetailsModal
