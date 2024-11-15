import { Flex, Link } from '@chakra-ui/react'
import React, { useEffect } from 'react'
import styles from './styles.module.scss'
import RMSwitch from '../../components/RMSwitch'
import { useDispatch } from 'react-redux'
import TableType2 from '../../components/TableType2'
import { switchGlobalSetting, setGlobalSetting } from '../../redux/globalClustersSlice'
import TextForm from '../../components/TextForm'
import { showErrorBanner } from '../../utility/common'

function CloudSettings({ config }) {
  const dispatch = useDispatch()
  const errInvalidGrant = (err) => { if (err?.message?.includes("invalid_grant")) err.message = <>{err.message}. <Link href="https://gitlab.signal18.io/users/sign_up" target='_blank'><u>Click here to Sign Up</u></Link></>; return err }

  useEffect(() => {
    // Re-render when the config prop changes
  }, [config]);

  const dataObject = [
    {
      key: 'Cloud18',
      value: (
        <RMSwitch
          confirmTitle={'Confirm switch global settings for cloud18?'}
          onChange={() => dispatch(switchGlobalSetting({ setting: 'cloud18', errMessage: errInvalidGrant }))}
          isChecked={config?.cloud18}
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
      key: 'Shared',
      value: (
        <RMSwitch
          confirmTitle={'Confirm switch global settings for shared cloud18?'}
          onChange={() => dispatch(switchGlobalSetting({ setting: 'cloud18Shared' }))}
          isChecked={config?.cloud18Shared}
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
    }
  ]

  return (
    <Flex justify='space-between' gap='0'>
      <TableType2 dataArray={dataObject} className={styles.table} />
    </Flex>
  )
}

export default CloudSettings
