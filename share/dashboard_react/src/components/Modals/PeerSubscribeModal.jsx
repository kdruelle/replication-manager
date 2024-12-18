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
import React, { useEffect, useState } from 'react'
import Markdown from 'react-markdown'
import RMButton from '../RMButton'
import { useTheme } from '../../ThemeProvider'
import parentStyles from './styles.module.scss'
import remarkGFM from 'remark-gfm'

function PeerSubscribeModal({ cluster, user, isOpen, closeModal, onSaveModal, terms = `` }) {
  const { theme } = useTheme()
  const [agreement, setAgreement] = useState(``)
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

  useEffect(() => {
    let servicePlan = Object.entries(cluster)
      .filter(([key]) => !([].includes(key))) // fields to remove
      .map(([key, value]) => `| ${key} | ${value} |`)
      .join("  \n");
    let finalterm = terms
      .replace(`<<user>>`, user?.username)
      .replace(`<<cluster>>`, clustername)
      .replace(`<<ervice_plan_infos>>`, "  \n".concat(servicePlan))
      .replace(`<<date>>`, (new Date()).toLocaleDateString())
    setAgreement(finalterm)
  }, [user, cluster, terms])

  return (
    <Modal size={'xl'} isOpen={isOpen} onClose={closeModal}>
      <ModalOverlay />
      <ModalContent className={theme === 'light' ? parentStyles.modalLightContent : parentStyles.modalDarkContent}>
        <ModalHeader>Terms and Conditions</ModalHeader>
        <ModalCloseButton />
        <ModalBody>
          <Stack spacing='5'>
            <Text>
              Cluster : {cluster?.["cluster-name"]}
            </Text>
            <Markdown  remarkPlugins={[remarkGFM]}>{agreement}</Markdown>
            <FormControl isInvalid={agreeError}>
              <Checkbox checked={agree} onCheckedChange={(e) => setAgree(!!e.checked)}>I agree with all terms and condition mentioned above</Checkbox>
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
