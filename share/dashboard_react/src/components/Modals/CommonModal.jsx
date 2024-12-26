import {
    Modal,
    ModalBody,
    ModalCloseButton,
    ModalContent,
    ModalHeader,
    ModalOverlay,
} from '@chakra-ui/react'
import React from 'react'
import parentStyles from './styles.module.scss'

function CommonModal({ body, title, size = 'md', isOpen, closeModal }) {
    return (
        <Modal size={size} isOpen={isOpen} onClose={closeModal}>
            <ModalOverlay />
            <ModalContent className={parentStyles.modalLightContent}>
                <ModalHeader>{title}</ModalHeader>
                <ModalCloseButton />
                <ModalBody>
                    {body}
                </ModalBody>
            </ModalContent>
        </Modal>
    )
}

export default CommonModal
