import { Flex, Link } from '@chakra-ui/react'
import React, { useEffect } from 'react'
import styles from './styles.module.scss'
import { useDispatch } from 'react-redux'
import TableType2 from '../../components/TableType2'
import { setGlobalSetting } from '../../redux/globalClustersSlice'
import LogSlider from '../../components/Sliders/LogSlider'

function ServerLogSettings({ config }) {
  const dispatch = useDispatch()
  
  useEffect(() => {
    // Re-render when the config prop changes
  }, [config]);

  const dataObject = [
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

export default ServerLogSettings
