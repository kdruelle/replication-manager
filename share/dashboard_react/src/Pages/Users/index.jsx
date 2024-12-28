import { createColumnHelper } from '@tanstack/react-table'
import React, { useEffect, useMemo, useState } from 'react'
import { DataTable } from '../../components/DataTable'
import AccordionComponent from '../../components/AccordionComponent'
import styles from './styles.module.scss'
import UserGrantModal from '../../components/Modals/UserGrantModal'
import { acceptSubscription, endSubscription, rejectSubscription } from '../../redux/clusterSlice'
import RMButton from '../../components/RMButton'
import RMIconButton from '../../components/RMIconButton'
import { HiUserGroup } from 'react-icons/hi'
import { TbUserCancel, TbUserStar } from 'react-icons/tb'
import ConfirmModal from '../../components/Modals/ConfirmModal'
import { useDispatch } from 'react-redux'
import { HStack } from '@chakra-ui/react'

function Users({ selectedCluster, user }) {
  const [data, setData] = useState([])
  const [selectedUser, setSelectedUser] = useState(null)
  const [action, setAction] = useState({ type: '', title: '', payload: '' })
  const [isUserGrantModalOpen, setIsUserGrantModalOpen] = useState(null)
  const [isConfirmModalOpen, setIsConfirmModalOpen] = useState(null)
  const columnHelper = createColumnHelper()
  const { type, title, payload } = action
  const dispatch = useDispatch()

  const showUser = (user, item) => {
    let normalUser = Object.entries(item.roles).every(([_, v]) => !v)

    if (user.user === "admin") {
      return true
    } else if (user.user === item.user) {
      return true
    } else if (user.roles['sysops']) {
      return true
    } else if (user.roles['dbops']) {
      return item.roles['extdbops']
    } else if (user.roles['sponsor']) {
      return item.roles['extdbops'] || item.roles['extsysops'] || normalUser
    }
    return false
  }

  useEffect(() => {
    if (selectedCluster?.apiUsers) {
      const result = Object.entries(selectedCluster?.apiUsers).filter(([_, value]) => showUser(user, value)).map(([key, value]) => ({
        user: key,
        ...value
      }));

      setData(result)
    }
  }, [selectedCluster?.apiUsers])

  const openUserGrantModal = () => {
    setIsUserGrantModalOpen(true)
  }

  const closeUserGrantModal = () => {
    setIsUserGrantModalOpen(false)
  }

  const openConfirmModal = () => {
    setIsConfirmModalOpen(true)
  }

  const closeConfirmModal = () => {
    setIsConfirmModalOpen(false)
  }

  const handleConfirm = () => {
    if (type === 'accept-sub') {
      dispatch(acceptSubscription({ clusterName: selectedCluster.name, username: payload }))
    } else if (type === 'reject-sub') {
      dispatch(rejectSubscription({ clusterName: selectedCluster.name, username: payload }))
    } else if (type === 'end-sub') {
      dispatch(endSubscription({ clusterName: selectedCluster.name, username: payload }))
    }
    closeConfirmModal()
  }

  const columns = useMemo(
    () => [
      columnHelper.accessor((row) => row.user, {
        cell: (info) => info.getValue(),
        header: 'User',
        id: 'user'
      }),
      columnHelper.accessor((row) => row.isExternal, {
        cell: (info) => info.getValue(),
        header: 'Is Git Validated',
        id: 'isExternal'
      }),
      columnHelper.accessor((row) => {
        return Object.entries(row.roles).filter(([_, v]) => v).map(([role, _]) => (<span>{role}</span>)).reduce((r, n) => {
          if (r.length) {
            r.push(<br />)
          }
          r.push(n)
          return r
        }, [])
      }, {
        cell: (info) => info.getValue(),
        header: 'Roles',
        id: 'roles'
      }),
      columnHelper.accessor((row) => (
        <HStack align={"center"} justifyContent={"center"}>
          { row?.roles?.["pending"] ? (
            <>
              <RMIconButton icon={TbUserStar} onClick={(e) => { e.stopPropagation(); setAction({ type: "accept-sub", title: "Are you sure to accept subscription?", payload: row.user }); openConfirmModal() }} />
              <RMIconButton icon={TbUserCancel} onClick={(e) => { e.stopPropagation(); setAction({ type: "reject-sub", title: "Are you sure to reject subscription?", payload: row.user }); openConfirmModal() }} />
            </>
          ) : (
            <>
              { row?.roles?.["sponsor"] && <RMIconButton icon={TbUserCancel} onClick={(e) => { e.stopPropagation(); setAction({ type: "end-sub", title: "Are you sure to end subscription?", payload: row.user }); openConfirmModal() }} />}
              <RMIconButton icon={HiUserGroup} onClick={(e) => { e.stopPropagation(); setSelectedUser(row); openUserGrantModal() }} />
            </>
          )}
        </HStack>
      ), {
        cell: (info) => info.getValue(),
        header: 'Grants',
        id: 'grants'
      })
    ],
    []
  )


  return (
    <>
      <AccordionComponent
        heading={'USERS'}
        allowToggle={false}
        className={styles.accordion}
        panelSX={{ overflowX: 'auto', p: 0 }}
        body={<DataTable data={data} columns={columns} className={styles.table} />}
      />
      {isUserGrantModalOpen && <UserGrantModal clusterName={selectedCluster.name} selectedUser={selectedUser} isOpen={isUserGrantModalOpen} closeModal={closeUserGrantModal} />}
      {isConfirmModalOpen && <ConfirmModal title={title} isOpen={isConfirmModalOpen} onConfirmClick={handleConfirm} closeModal={closeConfirmModal} />}
    </>
  )
}

export default Users
