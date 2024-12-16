import {
  Checkbox,
  FormControl,
  FormErrorMessage,
  FormLabel,
  Input,
  List,
  ListItem,
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalFooter,
  ModalHeader,
  ModalOverlay,
  Stack,
  VStack
} from '@chakra-ui/react'
import React, { useState, useEffect } from 'react'
import RMButton from '../RMButton'
import { useTheme } from '../../ThemeProvider'
import parentStyles from './styles.module.scss'
import { useDispatch, useSelector } from 'react-redux'
import { getMonitoredData } from '../../redux/globalClustersSlice'
import Message from '../Message'
import { updateGrants } from '../../redux/clusterSlice'
import GrantCheckList from '../GrantCheckList'

function UserGrantModal({ clusterName, user, isOpen, closeModal }) {
  const dispatch = useDispatch()

  const {
    globalClusters: { monitor }
  } = useSelector((state) => state)

  const [acls, setAcls] = useState([])
  const [grantsError, setGrantsError] = useState('')
  const [roles, setRoles] = useState([])
  const [rolesError, setRolesError] = useState('')
  const [allAcls, setAllAcls] = useState([])
  const [allRoles, setAllRoles] = useState([])
  const [firstLoad, setFirstLoad] = useState(true)
  const { theme } = useTheme()
  const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;

  useEffect(() => {
    if (monitor === null) {
      dispatch(getMonitoredData({}))
    }
  }, [monitor])

  useEffect(() => {
    setFirstLoad(true)
  }, [user])

   useEffect(() => {
    if (monitor?.serviceAcl?.length > 0 && firstLoad) {
      const modifiedWithSelectedProp = monitor?.serviceAcl.map((item) => Object.assign({}, item, { selected: user?.grants?.[item.grant] ?? false }))
      const modifiedRolesWithSelectedProp = monitor?.serviceRoles.map((item) => Object.assign({}, item, { selected: user?.roles?.[item.role] ?? false }))
      setAcls(modifiedWithSelectedProp)
      setAllAcls(modifiedWithSelectedProp)
      setRoles(modifiedRolesWithSelectedProp)
      setAllRoles(modifiedRolesWithSelectedProp)
      setFirstLoad(false)
    }
  }, [monitor?.serviceAcl, monitor?.serviceRoles, user])

  const handleCheckRoles = (e, role) => {
    const isChecked = e.target.checked;
    const updatedList = structuredClone(allRoles).map((x) => {
      if (x.role === role.role) {
        x.selected = isChecked;
      }
      return x;
    });
    setRoles(updatedList);
    setAllRoles(updatedList);
  };

  const handleSearchRoles = (e) => {
    const search = e.target.value
    if (search) {
      const searchRoleValue = search.toLowerCase()
      const searchedRoles = allRoles.filter((x) => {
        if (x.role.toLowerCase().includes(searchRoleValue)) {
          return x
        }
      })
      setRoles(searchedRoles)
    } else {
      setRoles(allRoles)
    }
  }

  const handleUserGrants = () => {
    setGrantsError('')
    setRolesError('')
  
    const selectedRoles = roles.filter((x) => x.selected).map((x) => x.role)
    if (selectedRoles.length === 0) {
      setRolesError('Please select at least one role')
      return
    }

    const selectedGrants = acls
    if (selectedGrants.length === 0) {
      setGrantsError('Please select at least one grant')
      return
    }

    dispatch(updateGrants({ clusterName, username: user?.user, grants: selectedGrants.join(' '), roles: selectedRoles.join(' ') }))
    closeModal()
  }

  return (
    <Modal isOpen={isOpen} onClose={closeModal}>
      <ModalOverlay />
      <ModalContent className={theme === 'light' ? parentStyles.modalLightContent : parentStyles.modalDarkContent}>
        <ModalHeader>{'Update user privileges'}</ModalHeader>
        <ModalCloseButton />
        <ModalBody>
          <Stack spacing='2'>
            <Message message={rolesError} />
            <VStack className={parentStyles.roleContainer}>
              <Input id='searchRole' type='search' onChange={handleSearchRoles} placeholder='Search ROLE' />
              <List className={parentStyles.roleList}>
                {roles.length > 0 &&
                  roles.map((role) => (
                    <ListItem className={parentStyles.roleListItem}>
                      <Checkbox
                        isChecked={!!roles.find((x) => x.role === role.role && x.selected)}
                        onChange={(e) => handleCheckRoles(e, role)}>
                        {role.role}
                      </Checkbox>
                    </ListItem>
                  ))}
              </List>
            </VStack>
            <Message message={grantsError} />
            <GrantCheckList grantOptions={allAcls} onChange={setAcls} parentStyles={parentStyles} />
          </Stack>
        </ModalBody>

        <ModalFooter gap={3} margin='auto'>
          <RMButton colorScheme='blue' size='medium' variant='outline' onClick={closeModal}>
            Cancel
          </RMButton>
          <RMButton onClick={handleUserGrants} size='medium'>
            Update Privileges
          </RMButton>
        </ModalFooter>
      </ModalContent>
    </Modal>
  )
}

export default UserGrantModal
