import {
  FormControl,
  FormErrorMessage,
  FormLabel,
  Input,
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalFooter,
  ModalHeader,
  ModalOverlay,
  Stack
} from '@chakra-ui/react'
import React, { useState } from 'react'
import RMButton from '../RMButton'
import { useTheme } from '../../ThemeProvider'
import parentStyles from './styles.module.scss'

function PeerLoginModal({ title, isOpen, closeModal, onSaveModal }) {
  const { theme } = useTheme()
  const [password, setPassword] = useState('')
  const [passwordError, setPasswordError] = useState('')

  const handleSave = () => {
    setPasswordError('')

    if (!password) {
      setPassword('Password is required')
      return
    }

    onSaveModal(password)
  }

  return (
    <Modal isOpen={isOpen} onClose={closeModal}>
      <ModalOverlay />
      <ModalContent className={theme === 'light' ? parentStyles.modalLightContent : parentStyles.modalDarkContent}>
        <ModalHeader>{title}</ModalHeader>
        <ModalCloseButton />
        <ModalBody>
          <Stack spacing='5'>
            <FormControl isInvalid={passwordError}>
              <FormLabel htmlFor='password'>Password</FormLabel>
              <Input
                id='password'
                type='password'
                isRequired={true}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
              <FormErrorMessage>{passwordError}</FormErrorMessage>
            </FormControl>
          </Stack>
        </ModalBody>

        <ModalFooter gap={3} margin='auto'>
          <RMButton colorScheme='blue' size='medium' variant='outline' onClick={closeModal}>
            Cancel
          </RMButton>
          <RMButton onClick={handleSave} size='medium'>
            Save
          </RMButton>
        </ModalFooter>
      </ModalContent>
    </Modal>
  )
}

export default PeerLoginModal
