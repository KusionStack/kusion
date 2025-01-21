import React, { useEffect, useRef, useState } from 'react'
import { Button, Card, message, Tabs, Tooltip, Modal } from 'antd'
import { useNavigate, useParams, useLocation } from 'react-router-dom'
import queryString from 'query-string'
import { StackService } from '@kusionstack/kusion-api-client-sdk'
import BackWithTitle from '@/components/backWithTitle'
import StackForm from '../components/stackForm'
import Runs from '../components/runs'
import ResourceGraph from '../components/resourceGraph'
import { KusionTabs } from '@/components/kusionTabs'

import styles from "./styles.module.less"


type TargetKey = React.MouseEvent | React.KeyboardEvent | string;

const tabsItems = [
  { label: 'Resource Graph', key: 'ResourceGraph' },
  { label: 'Runs', key: 'Runs' },
];

const ProjectDetail = () => {
  const navigate = useNavigate()
  const urlPrams = useParams()
  const location = useLocation()
  const funcRef = useRef<any>()
  const { projectName, stackId, panelKey } = queryString.parse(location?.search);

  const [activeKey, setActiveKey] = useState(stackId || undefined);
  const [items, setItems] = useState([]);
  const [stackFormOpen, setStackFormOpen] = useState(false)
  const [panelActiveKey, setPanelActiveKey] = useState(panelKey || tabsItems?.[0]?.key);
  const [formData, setFormData] = useState<any>()



  function handleStackTabChange(newActiveKey: string) {
    setActiveKey(newActiveKey);
    if (panelKey === 'Runs') {
      if (funcRef.current) {
        funcRef.current?.updateRunsURL({
          page: 1,
          stackId: newActiveKey
        })
      }
    } else {
      const newParams = queryString.stringify({
        projectName,
        stackId: newActiveKey,
        panelKey: 'ResourceGraph',
      })
      navigate(`?${newParams}`, { replace: true })
    }
  };

  function handlePanelTabChange(newActiveKey: string) {
    setPanelActiveKey(newActiveKey);
    const newParams = queryString.stringify({
      projectName,
      stackId: activeKey,
      panelKey: newActiveKey,
    })
    navigate(`?${newParams}`)
  };

  function remove(targetKey: TargetKey) {
    Modal.confirm({
      title: 'Are you sure to delete this stack?',
      okText: 'Yes',
      cancelText: 'No',
      onOk: async () => {
        const response = await StackService.deleteStack({
          path: {
            stackID: Number(targetKey),
          }
        })
        if (response?.data?.success) {
          message.success('Deleted Successful')
          getStackList({
            projectId: urlPrams?.projectId
          }, true)
        } else {
          message.error(response?.data?.message)
        }
      }
    });
  };

  function onEdit(action, key) {
    action === 'add' ? setStackFormOpen(true) : remove(key)
  }

  async function handleSubmit(values, callback) {
    let response: any
    if (formData?.id) {
      response = await StackService.updateStack({
        body: {
          ...values,
          projectID: Number(urlPrams?.projectId)
        },
        path: {
          stackID: formData?.id
        }
      })
    } else {
      response = await StackService.createStack({
        body: {
          ...values,
          projectID: Number(urlPrams?.projectId)
        }
      })
    }
    if (response?.data?.success) {
      message.success(formData?.id ? 'Update Successful' : 'Create Successful')
      callback && callback()
      getStackList({
        projectId: urlPrams?.projectId
      })
      handleClose()
    } else {
      message.error(response?.data?.message)
    }
  }
  function handleClose() {
    setStackFormOpen(false)
    setFormData(undefined)
  }

  async function getStackList(params, isDelete?: boolean) {
    try {
      const response: any = await StackService.listStack({
        query: {
          projectID: params?.projectId
        }
      });
      if (response?.data?.success) {
        const resTabs = response?.data?.data?.stacks?.map(item => {
          return {
            ...item,
            label: (
              <Tooltip title={`path: ${item?.path}`}>
                {item?.name}
              </Tooltip>
            ),
            key: item?.id,
          }
        })
        setItems(resTabs)
        setActiveKey(isDelete ? resTabs?.[0]?.key : (stackId || resTabs?.[0]?.key))
      } else {
        message.error(response?.data?.message)
      }
    } catch (error) {

    }
  }

  useEffect(() => {
    if (urlPrams?.projectId) {
      getStackList({
        projectId: urlPrams?.projectId
      })
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [urlPrams?.projectId])

  function handleBack() {
    navigate("/projects")
  }



  function handleClickEdit() {
    const currentStack = items?.find(item => Number(item?.id) === Number(activeKey))
    setFormData(currentStack)
    setStackFormOpen(true)
  }

  return (
    <div className={styles.project_detail}>
      <BackWithTitle title={projectName} handleBack={handleBack} />
      <Card>
        <div className={styles.project_detail_stackTab}>
          <KusionTabs
            items={items}
            addIsDiasble={false}
            activeKey={Number(activeKey) as any}
            handleClickItem={handleStackTabChange}
            onEdit={onEdit}
            disabledAdd={false}
          />
        </div>
        {
          activeKey && <>
            <div style={{ marginRight: 30 }}>
              <Tabs
                onChange={handlePanelTabChange}
                activeKey={panelActiveKey as string}
                items={tabsItems}
                tabBarExtraContent={<Button style={{ height: 45, marginBottom: 5, fontSize: 16 }} type="primary" onClick={handleClickEdit}>Edit Stack</Button>}
              />
            </div>
            {panelActiveKey === 'Runs' && <Runs ref={funcRef} stackId={activeKey} panelActiveKey={panelActiveKey} />}
            {panelActiveKey === 'ResourceGraph' && <ResourceGraph stackId={activeKey} />}
          </>
        }
        <StackForm
          formData={formData}
          stackFormOpen={stackFormOpen}
          handleCancel={handleClose}
          handleSubmit={handleSubmit}
        />
      </Card>
    </div>
  )
}

export default ProjectDetail
