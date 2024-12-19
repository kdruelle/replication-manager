import { Flex, Link } from '@chakra-ui/react'
import React, { useEffect } from 'react'
import styles from './styles.module.scss'
import { useDispatch } from 'react-redux'
import TableType2 from '../../components/TableType2'
import { setGlobalSetting, switchGlobalSetting } from '../../redux/globalClustersSlice'
import LogSlider from '../../components/Sliders/LogSlider'
import RMSwitch from '../../components/RMSwitch'
import TextForm from '../../components/TextForm'

function GlobalSettings({ config }) {
  const dispatch = useDispatch()
  
  useEffect(() => {
    // Re-render when the config prop changes
  }, [config]);

  const dataObject = [
    {
      key: 'API Public URL',
      value: (
        <TextForm
          value={config?.apiPublicUrl}
          confirmTitle={`Confirm API Public URL to `}
          onSave={(value) => {
            dispatch(setGlobalSetting({ setting: 'api-public-url', value }))
          }}
        />
      )
    },
    {
      key: 'Mail From',
      value: (
        <TextForm
          value={config?.mailFrom}
          confirmTitle={`Confirm mail-from to `}
          onSave={(value) => {
            dispatch(setGlobalSetting({ setting: 'mail-from', value }))
          }}
        />
      )
    },
    {
      key: 'Mail To',
      value: (
        <TextForm
          value={config?.mailTo}
          confirmTitle={`Confirm mail-to to `}
          onSave={(value) => {
            dispatch(setGlobalSetting({ setting: 'mail-to', value }))
          }}
        />
      )
    },
    {
      key: 'Mail SMTP Address',
      value: (
        <TextForm
          value={config?.mailSmtpAddr}
          confirmTitle={`Confirm Mail SMTP Address to `}
          onSave={(value) => {
            dispatch(setGlobalSetting({ setting: 'mail-smtp-addr', value }))
          }}
        />
      )
    },
    {
      key: 'Mail SMTP User',
      value: (
        <TextForm
          value={config?.mailSmtpUser}
          confirmTitle={`Confirm Mail SMTP User to `}
          onSave={(value) => {
            dispatch(setGlobalSetting({ setting: 'mail-smtp-user', value }))
          }}
        />
      )
    },
    {
      key: 'Mail SMTP Password',
      value: (
        <TextForm
          value={config?.mailSmtpPassword}
          confirmTitle={`Confirm Mail SMTP Password to `}
          onSave={(value) => {
            dispatch(setGlobalSetting({ setting: 'mail-smtp-password', value: btoa(value) }))
          }}
        />
      )
    },
    {
      key: 'Mail SMTP TLS (Skip Verify)',
      value: (
        <RMSwitch
          confirmTitle={'Confirm switch global settings for Mail SMTP TLS?'}
          onChange={(_v, setRefresh) => dispatch(switchGlobalSetting({ setting: 'mail-smtp-tls-skip-verify', errMessage: errInvalidGrant, setRefresh }))}
          isChecked={config?.mailSmtpTlsSkipVerify}
        />
      )
    },
    {
      key: 'Log File Level',
      value: (
        <LogSlider
          value={config?.logFileLevel}
          confirmTitle={`Confirm change 'log-file-level' to: `}
          onChange={(val) =>
            dispatch(
              setGlobalSetting({
                setting: 'log-file-level',
                value: val
              })
            )
          }
        />
      )
    },
    {
      key: 'Log GIT',
      value: (
        <LogSlider
          value={config?.logGitLevel}
          confirmTitle={`Confirm change 'log-git-level' to: `}
          onChange={(val) =>
            dispatch(
              setGlobalSetting({
                setting: 'log-git-level',
                value: val
              })
            )
          }
        />
      )
    },
  ]

  return (
    <Flex justify='space-between' gap='0'>
      <TableType2 dataArray={dataObject} className={styles.table} />
    </Flex>
  )
}

export default GlobalSettings
