import { Flex, Link } from '@chakra-ui/react'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import styles from './styles.module.scss'
import RMSwitch from '../../components/RMSwitch'
import { useDispatch } from 'react-redux'
import TableType2 from '../../components/TableType2'
import { switchGlobalSetting, setGlobalSetting, reloadClustersPlan } from '../../redux/globalClustersSlice'
import TextForm from '../../components/TextForm'
import RMIconButton from '../../components/RMIconButton'
import { HiRefresh } from 'react-icons/hi'
import ConfirmModal from '../../components/Modals/ConfirmModal'
import TagPill from '../../components/TagPill'
import RMButton from '../../components/RMButton'

function CloudSettings({ config }) {
  const dispatch = useDispatch()
  const [action, setAction] = useState({
    title: '',
    type: '',
  })
  const {title,type} = action
  const errInvalidGrant = (err) => { if (err?.message?.includes("invalid_grant")) err.message = <>{err.message}. <Link href="https://gitlab.signal18.io/users/sign_up" target='_blank'><u>Click here to Sign Up</u></Link></>; return err }


  const openConfirmModal = () => {
    setIsConfirmModalOpen(true)
  }

  const closeConfirmModal = () => {
    setIsConfirmModalOpen(false)
  }

  const actionHandler = useCallback(() => {
    if (type === 'cloud18-connect') {
      dispatch(setGlobalSetting({ setting: 'cloud18', value: "true", errMsgFunc: errInvalidGrant }))
    } else if (type === 'cloud18-disconnect') {
      dispatch(setGlobalSetting({ setting: 'cloud18', value: "false", errMsgFunc: errInvalidGrant }))
    } else if (type === 'reload-clusters-plan') {
      dispatch(reloadClustersPlan({}))
    }
  }, [type])

  const disableConnect = useMemo(() => (config?.cloud18GitUser === "" || config?.cloud18Domain === "" || config?.cloud18SubDomain === "" || config?.cloud18SubDomainZone === ""),[config?.cloud18GitUser, config?.cloud18Domain, config?.cloud18SubDomain, config?.cloud18SubDomainZone])

  const [isConfirmModalOpen, setIsConfirmModalOpen] = useState(false)

  useEffect(() => {
    // Re-render when the config prop changes
  }, [config]);

  const dataObject = [
    {
      key: 'Cloud18 Status',
      value: (
        <TagPill colorScheme={config?.cloud18 ? 'green' : 'gray'} text={config?.cloud18 ? 'ONLINE' : 'OFFLINE'} />
      )
    },
    {
      key: 'Git user',
      value: (
        <TextForm
          value={config?.cloud18GitUser}
          confirmTitle={`Confirm git username to `}
          onSave={(value) => {
            dispatch(setGlobalSetting({ setting: 'cloud18-gitlab-user', value }))
          }}
        />
      )
    },
    {
      key: 'Gitlab Password',
      value: (
        <TextForm
          value={config?.cloud18GitlabPassword}
          confirmTitle={`Confirm gitlab password to `}
          onSave={(value) => {
            dispatch(setGlobalSetting({ setting: 'cloud18-gitlab-password', value: btoa(value) }))
          }}
        />
      )
    },
    {
      key: 'Domain',
      value: (
        <TextForm
          value={config?.cloud18Domain}
          confirmTitle={`Confirm cloud18 Domain to `}
          onSave={(value) => {
            dispatch(setGlobalSetting({ setting: 'cloud18-domain', value }))
          }}
        />
      )
    },
    {
      key: 'Subdomain', 
      value: (
        <TextForm
          value={config?.cloud18SubDomain}
          confirmTitle={`Confirm subdomain to `}
          onSave={(value) => {
            dispatch(setGlobalSetting({ setting: 'cloud18-sub-domain', value }))
          }}
        />
      )
    },
    {
      key: 'Subdomain zone',
      value: (
        <TextForm
          value={config?.cloud18SubDomainZone}
          confirmTitle={`Confirm subdomain zone to `}
          onSave={(value) => {
            dispatch(setGlobalSetting({ setting: 'cloud18-sub-domain-zone', value }))
          }}
        />
      )
    },
    {
      key: 'Platform Description',
      value: (
        <TextForm
          value={config?.cloud18PlatformDescription}
          confirmTitle={`Confirm platform description to `}
          onSave={(value) => {
            dispatch(setGlobalSetting({ setting: 'cloud18-platform-description', value }))
          }}
        />
      )
    },
    {
      key: 'Cloud18 Connect',
      value: (<> { config?.cloud18 ? <RMButton onClick={() => { setAction({title:'Confirm disconnect from cloud18?', type: 'cloud18-disconnect'}); openConfirmModal()}}>Disconnect</RMButton> : <RMButton isDisabled={disableConnect}  onClick={() => { setAction({title:'Confirm connect to cloud18?', type: 'cloud18-connect'}); openConfirmModal()}}>Connect</RMButton>}</>)
    },
    {
      key: 'Reload All Clusters Plans',
      value: (
        <RMIconButton icon={HiRefresh} onClick={() => { setAction({title:'Confirm reload all clusters plans?', type: 'reload-clusters-plan'}); openConfirmModal() }} />
      )
    },
  ]

  return (
    <Flex justify='space-between' gap='0'>
      <TableType2 dataArray={dataObject} className={styles.table} />
      {isConfirmModalOpen && (
        <ConfirmModal
          isOpen={isConfirmModalOpen}
          closeModal={() => {
            closeConfirmModal()
          }}
          title={title}
          onConfirmClick={() => {
            actionHandler()
            closeConfirmModal()
          }}
        />
      )}
    </Flex>
  )
}

export default CloudSettings
