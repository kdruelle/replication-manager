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
import { addUser } from '../../redux/clusterSlice'

function AddUserModal({ clusterName, isOpen, closeModal }) {
  const dispatch = useDispatch()

  const {
    globalClusters: { monitor }
  } = useSelector((state) => state)

  const [userName, setUserName] = useState('')
  const [userNameError, setUserNameError] = useState('')
  // const [password, setPassword] = useState('')
  // const [passwordError, setPasswordError] = useState('')
  const [acls, setAcls] = useState([])
  const [grantsError, setGrantsError] = useState('')
  const [roles, setRoles] = useState([])
  const [rolesError, setRolesError] = useState('')
  const [allAcls, setAllAcls] = useState([])
  const [allRoles, setAllRoles] = useState([])
  const [firstLoad, setFirstLoad] = useState(true)
  const { theme } = useTheme()
  const emailRegex = /^[a-zA-Z0-9](\.?[a-zA-Z0-9]){5,50}@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
  /*
    Usage examples
    emailRegex.test("first.last@domain.com");      // true
    emailRegex.test("firstname@domain.com");       // true
    emailRegex.test("user123@domain.co.uk");       // true
    emailRegex.test("first..last@domain.com");     // false, consecutive dots
    emailRegex.test("plainaddress");               // false, no domain
    emailRegex.test("missing@dotcom");             // false, invalid TLD
  */

  useEffect(() => {
    if (monitor === null) {
      dispatch(getMonitoredData({}))
    }
  }, [monitor])

  useEffect(() => {
    if (monitor?.serviceAcl?.length > 0 && firstLoad) {
      const modifiedWithSelectedProp = monitor?.serviceAcl.map((item) => Object.assign({}, item, { selected: false }))
      const modifiedRolesWithSelectedProp = monitor?.serviceRoles.map((item) => Object.assign({}, item, { selected: false }))
      setAcls(modifiedWithSelectedProp)
      setAllAcls(modifiedWithSelectedProp)
      setRoles(modifiedRolesWithSelectedProp)
      setAllRoles(modifiedRolesWithSelectedProp)
      setFirstLoad(false)
    }
  }, [monitor?.serviceAcl, monitor?.serviceRoles])

  const handleCheck = (e, acl) => {
    const isChecked = e.target.checked
    const updatedList = allAcls.map((x) => {
      if (x.grant === acl.grant) {
        x.selected = isChecked
      }
      return x
    })
    setAcls(updatedList)
    setAllAcls(updatedList)
  }

  const handleCheckRoles = (e, role) => {
    const isChecked = e.target.checked
    const updatedList = allRoles.map((x) => {
      if (x.role === role.role) {
        x.selected = isChecked
      }
      return x
    })
    setRoles(updatedList)
    setAllRoles(updatedList)
  }

  const handleSearch = (e) => {
    const search = e.target.value
    if (search) {
      const searchValue = search.toLowerCase()
      const searchedAcls = allAcls.filter((x) => {
        if (x.grant.toLowerCase().includes(searchValue)) {
          return x
        }
      })
      setAcls(searchedAcls)
    } else {
      setAcls(allAcls)
    }
  }

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

  const handleAddUser = () => {
    setUserNameError('')
    // setPasswordError('')
    setGrantsError('')
    setRolesError('')
    if (!userName) {
      setUserNameError('User is required')
      return
    }

    if (!emailRegex.test(userName)) {
      setUserNameError('User must be email address')
      return
    }

    const selectedRoles = roles.filter((x) => x.selected).map((x) => x.role)
    if (selectedRoles.length === 0) {
      setRolesError('Please select atleast one role')
      return
    }
    
    const selectedGrants = acls.filter((x) => x.selected).map((x) => x.grant)
    if (selectedGrants.length === 0) {
      setGrantsError('Please select atleast one grant')
      return
    }

    dispatch(addUser({ clusterName, username: userName, grants: selectedGrants.join(' '), roles: selectedRoles.join(' ') }))
    closeModal()
  }
  return (
    <Modal isOpen={isOpen} onClose={closeModal}>
      <ModalOverlay />
      <ModalContent className={theme === 'light' ? parentStyles.modalLightContent : parentStyles.modalDarkContent}>
        <ModalHeader>{'Add a new user'}</ModalHeader>
        <ModalCloseButton />
        <ModalBody>
          <Stack spacing='2'>
            <FormControl isInvalid={userNameError}>
              <FormLabel htmlFor='username'>User Email</FormLabel>
              <Input
                id='username'
                type='text'
                isRequired={true}
                value={userName}
                onChange={(e) => setUserName(e.target.value)}
              />
              <FormErrorMessage>{userNameError}</FormErrorMessage>
            </FormControl>
            {/* <FormControl isInvalid={passwordError}>
              <FormLabel htmlFor='password'>Password</FormLabel>
              <Input id='password' type='password' value={password} onChange={(e) => setPassword(e.target.value)} />
              <Message type='error' message={passwordError} />
            </FormControl> */}
            <Message message={rolesError} />
            <VStack className={parentStyles.roleContainer}>
              <Input id='searchRole' type='search' onChange={handleSearchRoles} placeholder='Search ROLE' />
              <List className={parentStyles.roleList}>
                {roles.length > 0 &&
                  roles.map((role) => (
                    <ListItem className={parentStyles.roleListItem}>
                      <Checkbox
                        size='lg'
                        isChecked={!!roles.find((x) => x.role === role.role && x.selected)}
                        onChange={(e) => handleCheckRoles(e, role)}>
                        {role.role}
                      </Checkbox>
                    </ListItem>
                  ))}
              </List>
            </VStack>
            <Message message={grantsError} />
            <VStack className={parentStyles.aclContainer}>
              <Input id='searchAcl' type='search' onChange={handleSearch} placeholder='Search ACL' />
              <List className={parentStyles.aclList}>
                {acls.length > 0 &&
                  acls.map((acl) => (
                    <ListItem className={parentStyles.aclListItem}>
                      <Checkbox
                        size='lg'
                        isChecked={!!acls.find((x) => x.grant === acl.grant && x.selected)}
                        onChange={(e) => handleCheck(e, acl)}>
                        {acl.grant}
                      </Checkbox>
                    </ListItem>
                  ))}
              </List>
            </VStack>
          </Stack>
        </ModalBody>

        <ModalFooter gap={3} margin='auto'>
          <RMButton colorScheme='blue' size='medium' variant='outline' onClick={closeModal}>
            Cancel
          </RMButton>
          <RMButton onClick={handleAddUser} size='medium'>
            Add User
          </RMButton>
        </ModalFooter>
      </ModalContent>
    </Modal>
  )
}

export default AddUserModal
