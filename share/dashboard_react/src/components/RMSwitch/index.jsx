import { Flex, Spinner, Switch, Text } from '@chakra-ui/react'
import React, { useState, useEffect } from 'react'
import styles from './styles.module.scss'
import ConfirmModal from '../Modals/ConfirmModal'

function RMSwitch({
  id,
  onText = 'ON',
  offText = 'OFF',
  isChecked,
  size = 'md',
  isDisabled,
  onChange,
  confirmTitle,
  loading
}) {
  const [currentValue, setCurrentValue] = useState(isChecked)
  const [previousValue, setPreviousValue] = useState(isChecked)
  const [refresh, setRefresh] = useState("")
  const [isConfirmModalOpen, setIsConfirmModalOpen] = useState(false)

  useEffect(() => {
    setCurrentValue(isChecked)
    setPreviousValue(isChecked)
  }, [refresh, isChecked])

  const handleChange = (e) => {
    setCurrentValue(e.target.checked)
    if (confirmTitle) {
      openConfirmModal()
    } else {
      onChange(e.target.checked,setRefresh)
    }
  }

  const openConfirmModal = () => {
    setIsConfirmModalOpen(true)
  }
  const closeConfirmModal = (action) => {
    if (action === 'cancel') {
      setCurrentValue(previousValue)
    }
    setIsConfirmModalOpen(false)
  }

  return (
    <Flex className={styles.switchContainer} align='center'>
      <Switch key={refresh} size={size} id={id} isChecked={currentValue} isDisabled={isDisabled} onChange={handleChange} />
      <Text className={`${styles.text} ${currentValue ? styles.green : styles.red}`}>
        {currentValue ? onText : offText}
      </Text>
      {loading && <Spinner />}
      {isConfirmModalOpen && (
        <ConfirmModal
          isOpen={isConfirmModalOpen}
          closeModal={() => {
            closeConfirmModal('cancel')
          }}
          title={confirmTitle}
          onConfirmClick={() => {
            onChange(currentValue,setRefresh)
            closeConfirmModal('')
          }}
        />
      )}
    </Flex>
  )
}

export default RMSwitch
