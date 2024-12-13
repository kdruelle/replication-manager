import { Flex } from '@chakra-ui/react'
import React, { useState, useEffect } from 'react'
import styles from './styles.module.scss'
import RMSwitch from '../../components/RMSwitch'
import { useDispatch, useSelector } from 'react-redux'
import TableType2 from '../../components/TableType2'
import Dropdown from '../../components/Dropdown'
import { convertObjectToArrayForDropdown, formatBytes } from '../../utility/common'
import { setSetting, switchSetting } from '../../redux/settingsSlice'
import TextForm from '../../components/TextForm'

function CloudSettings({ selectedCluster, user }) {
  const dispatch = useDispatch()

  const {
    globalClusters: { monitor }
  } = useSelector((state) => state)

  const [planOptions, setPlanOptions] = useState([])
  const getPlanOptions = (plist = []) => [{ name: "No Plan", value: '' }, ...plist?.map((obj) => ({ name: obj.plan, value: obj.plan }))]

  const onPlanChange = (option) => {
    setPlan(option.value)
    //setPlanDetails(planOptions.find((obj) => obj.plan === option.value))
    //console.log(plans,option)
  }
  useEffect(() => {
    if (monitor?.servicePlans) {
      setPlanOptions(getPlanOptions(monitor.servicePlans))
    }
  }, [monitor?.servicePlans])

  const dataObject = [
    ...(selectedCluster?.config?.cloud18
      ? [
        {
          key: 'For Sale',
          value: (
            <RMSwitch
              confirmTitle={'Confirm switch settings for cloud18-shared?'}
              onChange={() =>
                dispatch(switchSetting({ clusterName: selectedCluster?.name, setting: 'cloud18-shared' }))
              }
              isDisabled={user?.grants['cluster-settings'] == false}
              isChecked={selectedCluster?.config?.cloud18Shared}
            />
          )
        },
        {
          key: 'Cluster Plan',
          value: (
            <Flex className={styles.dropdownContainer}>
              <Dropdown
                options={planOptions}
                id='plan'
                className={styles.dropdownButton}
                selectedValue={selectedCluster?.config?.provServicePlan}
                confirmTitle={`Confirm plan change to`}
                onChange={(option) => {
                  dispatch(
                    setSetting({
                      clusterName: selectedCluster?.name,
                      setting: 'prov-service-plan',
                      value: option
                    })
                  )
                }}
              />
            </Flex>
          )
        },
        {
          key: 'Cloud18 Database Read-Write-Split Srv Record',
          value: (
            <TextForm
              value={selectedCluster?.config?.cloud18DatabaseReadWriteSplitSrvRecord}
              confirmTitle={`Confirm cloud18-database-read-write-split-srv-record to `}
              maxLength={1024}
              className={styles.textbox}
              onSave={(value) =>
                dispatch(
                  setSetting({
                    clusterName: selectedCluster?.name,
                    setting: 'cloud18-database-read-write-split-srv-record',
                    value: value
                  })
                )
              }
            />
          )
        },
        {
          key: 'Cloud18 Database Read-Write Srv Record',
          value: (
            <TextForm 
              value={selectedCluster?.config?.cloud18DatabaseReadWriteSrvRecord}
              confirmTitle={`Confirm cloud18-database-read-write-srv-record to `}
              maxLength={1024}
              className={styles.textbox}
              onSave={(value) =>
                dispatch(
                  setSetting({
                    clusterName: selectedCluster?.name,
                    setting: 'cloud18-database-read-write-srv-record',
                    value: value
                  })
                )
              }
            />
          )
        },
        {
          key: 'Cloud18 Database Read Srv Record',
          value: (
            <TextForm
              value={selectedCluster?.config?.cloud18DatabaseReadSrvRecord}
              confirmTitle={`Confirm cloud18-database-read-srv-record to `}
              maxLength={1024}
              className={styles.textbox}
              onSave={(value) =>
                dispatch(
                  setSetting({
                    clusterName: selectedCluster?.name,
                    setting: 'cloud18-database-read-srv-record',
                    value: value
                  })
                )
              }
            />
          )
        }
      ]
      : [])
  ]

  return (
    <Flex justify='space-between' gap='0'>
      <TableType2 dataArray={dataObject} className={styles.table} />
    </Flex>
  )
}

export default CloudSettings
