import {
  Checkbox,
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
  Stack,
  Text
} from '@chakra-ui/react'
import React, { useState } from 'react'
import RMButton from '../RMButton'
import { useTheme } from '../../ThemeProvider'
import parentStyles from './styles.module.scss'

function PeerSubscribeModal({ cluster, isOpen, closeModal, onSaveModal }) {
  const { theme } = useTheme()
  const [agree, setAgree] = useState(false)
  const [agreeError, setAgreeError] = useState('')

  const handleSubmit = () => {
    setAgreeError('')

    if (!agree) {
      setAgreeError("You need to accept terms and condition to use the cluster")
      return
    }

    onSaveModal(cluster)
  }

  return (
    <Modal isOpen={isOpen} onClose={closeModal}>
      <ModalOverlay />
      <ModalContent className={theme === 'light' ? parentStyles.modalLightContent : parentStyles.modalDarkContent}>
        <ModalHeader>Terms and Conditions</ModalHeader>
        <ModalCloseButton />
        <ModalBody>
          <Stack spacing='5'>
            <Text>
              Cluster : {cluster?.["cluster-name"]}
            </Text>
            <Text>
              Lorem ipsum dolor sit amet, consectetur adipiscing elit. Curabitur sodales
              nisl eget elit vestibulum, at suscipit justo efficitur. Donec elementum
              mauris eget risus fermentum luctus. Pellentesque habitant morbi tristique
              senectus et netus et malesuada fames ac turpis egestas. Sed volutpat velit
              sit amet eros sodales dignissim. Vivamus in consectetur mauris. Fusce non
              enim a risus malesuada placerat sed vitae leo. Nunc vehicula erat vel risus
              bibendum, at tincidunt velit gravida.
            </Text>
            <FormControl isInvalid={agreeError}>
              <Checkbox checked={agree} onCheckedChange={(e) => setChecked(!!e.checked)}>I agree with all terms and condition mentioned above</Checkbox>
              <FormErrorMessage>{agreeError}</FormErrorMessage>
            </FormControl>
          </Stack>
        </ModalBody>

        <ModalFooter gap={3} margin='auto'>
          <RMButton colorScheme='blue' size='medium' variant='outline' onClick={closeModal}>
            Cancel
          </RMButton>
          <RMButton onClick={handleSubmit} size='medium'>
            Submit
          </RMButton>
        </ModalFooter>
      </ModalContent>
    </Modal>
  )
}

export default PeerSubscribeModal
