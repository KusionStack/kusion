import React, { forwardRef, useImperativeHandle, useRef } from 'react'
import G6 from '@antv/g6'
import type {
  IG6GraphEvent,
  IGroup,
  ModelConfig,
  IAbstractGraph,
} from '@antv/g6'
import { useLocation } from 'react-router-dom'
import queryString from 'query-string'
import Loading from '@/components/loading'
import { ICON_MAP } from '@/utils/images'

import styles from './style.module.less'
interface NodeConfig extends ModelConfig {
  data?: {
    name?: string
    count?: number
    resourceGroup?: {
      name: string
      [key: string]: any
    }
  }
  label?: string
  id?: string
  resourceGroup?: {
    name: string
    [key: string]: any
  }
}

interface NodeModel {
  id: string
  name?: string
  label?: string
  resourceGroup?: {
    name: string
  }
  data?: {
    count?: number
    resourceGroup?: {
      name: string
    }
  }
}

function getTextWidth(str: string, fontSize: number) {
  const canvas = document.createElement('canvas')
  const context = canvas.getContext('2d')!
  context.font = `${fontSize}px sans-serif`
  return context.measureText(str).width
}

function fittingString(str: string, maxWidth: number, fontSize: number) {
  const ellipsis = '...'
  const ellipsisLength = getTextWidth(ellipsis, fontSize)

  if (maxWidth <= 0) {
    return ''
  }

  const width = getTextWidth(str, fontSize)
  if (width <= maxWidth) {
    return str
  }

  let len = str.length
  while (len > 0) {
    const substr = str.substring(0, len)
    const subWidth = getTextWidth(substr, fontSize)

    if (subWidth + ellipsisLength <= maxWidth) {
      return substr + ellipsis
    }

    len--
  }

  return str
}

function getNodeName(cfg: NodeConfig, type: string) {
  if (type === 'resource') {
    const [left, right] = cfg?.id?.split(':') || []
    const leftList = left?.split('.')
    const leftListLength = leftList?.length || 0
    const leftLast = leftList?.[leftListLength - 1]
    return `${leftLast}:${right}`
  }
  const list = cfg?.label?.split('.')
  const len = list?.length || 0
  return list?.[len - 1] || ''
}

type IProps = {
  topologyLoading?: boolean
  onTopologyNodeClick?: (node: any) => void
  isResource?: boolean
  tableName?: string
  handleChangeCluster?: (val: any) => void
  selectedCluster?: string
  clusterOptions?: string[]
}

const TopologyMap = forwardRef((props: IProps, drawRef) => {
  const {
    onTopologyNodeClick,
    topologyLoading,
    isResource,
    tableName,
  } = props
  const containerRef = useRef(null)
  const graphRef = useRef<IAbstractGraph | null>(null)
  const location = useLocation()
  const { type } = queryString.parse(location?.search)


  function handleMouseEnter(evt) {
    graphRef.current?.setItemState(evt.item, 'hoverState', true)
  }

  const handleMouseLeave = (evt: IG6GraphEvent) => {
    graphRef.current?.setItemState(evt.item, 'hoverState', false)
  }

  G6.registerNode(
    'card-node',
    {
      draw(cfg: NodeConfig, group: IGroup) {
        const displayName = getNodeName(cfg, type as string)
        const count = cfg.data?.count
        const nodeWidth = type === 'cluster' ? 240 : 200

        // Create main container
        const rect = group.addShape('rect', {
          attrs: {
            x: 0,
            y: 0,
            width: nodeWidth,
            height: 48,
            radius: 6,
            fill: '#ffffff',
            stroke: '#e6f4ff',
            lineWidth: 1,
            shadowColor: 'rgba(0,0,0,0.06)',
            shadowBlur: 8,
            shadowOffsetX: 0,
            shadowOffsetY: 2,
            cursor: 'pointer',
          },
          name: 'node-container',
        })

        // Add side accent
        group.addShape('rect', {
          attrs: {
            x: 0,
            y: 0,
            width: 3,
            height: 48,
            radius: [3, 0, 0, 3],
            fill: '#1677ff',
            opacity: 0.4,
          },
          name: 'node-accent',
        })

        // Add Kubernetes icon
        const iconSize = 32
        const kind = cfg?.data?.resourceGroup?.kind || ''
        group.addShape('image', {
          attrs: {
            x: 16,
            y: (48 - iconSize) / 2,
            width: iconSize,
            height: iconSize,
            img: ICON_MAP[kind as keyof typeof ICON_MAP] || ICON_MAP.Kubernetes,
          },
          name: 'node-icon',
        })

        // Add title text
        group.addShape('text', {
          attrs: {
            x: 52,
            y: 24,
            text: fittingString(displayName || '', 100, 14),
            fontSize: 14,
            fontWeight: 500,
            fill: '#1677ff',
            cursor: 'pointer',
            textBaseline: 'middle',
            fontFamily:
              '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial',
          },
          name: 'node-label',
        })

        if (typeof count === 'number') {
          const textWidth = getTextWidth(`${count}`, 12)
          const circleSize = Math.max(textWidth + 12, 20)
          const circleX = 170
          const circleY = 24

          // Add count background
          group.addShape('circle', {
            attrs: {
              x: circleX,
              y: circleY,
              r: circleSize / 2,
              fill: '#f0f5ff',
            },
            name: 'count-background',
          })

          // Add count text
          group.addShape('text', {
            attrs: {
              x: circleX,
              y: circleY,
              text: `${count}`,
              fontSize: 12,
              fontWeight: 500,
              fill: '#1677ff',
              textAlign: 'center',
              textBaseline: 'middle',
              fontFamily:
                '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial',
            },
            name: 'count-text',
          })
        }

        return rect
      },
    },
    'single-node',
  )

  G6.registerEdge(
    'running-edge',
    {
      afterDraw(cfg, group) {
        const shape = group?.get('children')[0]
        if (!shape) return

        // Get the path shape
        const startPoint = shape.getPoint(0)

        // Create animated circle
        const circle = group.addShape('circle', {
          attrs: {
            x: startPoint.x,
            y: startPoint.y,
            fill: '#1677ff',
            r: 2,
            opacity: 0.8,
          },
          name: 'running-circle',
        })

        // Add movement animation
        circle.animate(
          ratio => {
            const point = shape.getPoint(ratio)
            return {
              x: point.x,
              y: point.y,
            }
          },
          {
            repeat: true,
            duration: 2000,
          },
        )
      },
      setState(name, value, item) {
        const shape = item.get('keyShape')
        if (name === 'hover') {
          shape?.attr('stroke', value ? '#1677ff' : '#c2c8d1')
          shape?.attr('lineWidth', value ? 2 : 1)
          shape?.attr('strokeOpacity', value ? 1 : 0.7)
        }
      },
    },
    'cubic', // Extend from built-in cubic edge
  )

  function initGraph() {
    const container = containerRef.current
    const width = container?.scrollWidth
    const height = container?.scrollHeight
    const toolbar = new G6.ToolBar()
    return new G6.Graph({
      container,
      width,
      height,
      fitCenter: true,
      plugins: [toolbar],
      enabledStack: true,
      modes: {
        default: ['drag-canvas', 'drag-node', 'click-select'],
      },
      animate: true,
      layout: {
        type: 'dagre',
        rankdir: 'LR',
        align: 'UL',
        nodesep: 10,
        ranksep: 40,
        nodesepFunc: () => 1,
        ranksepFunc: () => 1,
        controlPoints: true,
        sortByCombo: false,
        preventOverlap: true,
        nodeSize: [200, 60],
        workerEnabled: true,
        clustering: false,
        clusterNodeSize: [200, 60],
        // Optimize edge layout
        edgeFeedbackStyle: {
          stroke: '#c2c8d1',
          lineWidth: 1,
          strokeOpacity: 0.5,
          endArrow: true,
        },
      },
      defaultNode: {
        type: 'card-node',
        size: [200, 60],
        style: {
          fill: '#fff',
          stroke: '#e5e6e8',
          radius: 4,
          shadowColor: 'rgba(0,0,0,0.05)',
          shadowBlur: 4,
          shadowOffsetX: 0,
          shadowOffsetY: 2,
          cursor: 'pointer',
        },
      },
      defaultEdge: {
        type: 'running-edge',
        style: {
          radius: 10,
          offset: 5,
          endArrow: {
            path: G6.Arrow.triangle(4, 6, 0),
            d: 0,
            fill: '#c2c8d1',
          },
          stroke: '#c2c8d1',
          lineWidth: 1,
          strokeOpacity: 0.7,
          curveness: 0.5,
        },
        labelCfg: {
          autoRotate: true,
          style: {
            fill: '#86909c',
            fontSize: 12,
          },
        },
      },
      edgeStateStyles: {
        hover: {
          lineWidth: 2,
        },
      },
      nodeStateStyles: {
        selected: {
          stroke: '#1677ff',
          shadowColor: 'rgba(22,119,255,0.12)',
          fill: '#f0f5ff',
          opacity: 0.8,
        },
        hoverState: {
          stroke: '#1677ff',
          shadowColor: 'rgba(22,119,255,0.12)',
          fill: '#f0f5ff',
          opacity: 0.8,
        },
        clickState: {
          stroke: '#1677ff',
          shadowColor: 'rgba(22,119,255,0.12)',
          fill: '#f0f5ff',
          opacity: 0.8,
        },
      },
    })
  }

  function setHightLight() {
    graphRef.current.getNodes().forEach(node => {
      const model: any = node.getModel()
      const displayName = getNodeName(model, type as string)
      const isHighLight =
        type === 'resource'
          ? model?.data?.resourceGroup?.name === tableName
          : displayName === tableName
      if (isHighLight) {
        graphRef.current?.setItemState(node, 'selected', true)
      }
    })
  }

  function drawGraph(topologyData) {
    if (topologyData) {
      if (type === 'resource') {
        graphRef.current?.destroy()
        graphRef.current = null
      }
      if (!graphRef.current) {
        graphRef.current = initGraph()
        graphRef.current?.read(topologyData)

        setHightLight()

        graphRef.current?.on('node:click', evt => {
          const node = evt.item
          const model = node.getModel()
          graphRef.current?.getNodes().forEach(n => {
            graphRef.current?.setItemState(n, 'selected', false)
          })
          graphRef.current?.setItemState(node, 'selected', true)
          onTopologyNodeClick?.(model)
        })

        graphRef.current?.on('node:mouseenter', evt => {
          const node = evt.item
          if (
            !graphRef.current
              ?.findById(node.getModel().id)
              ?.hasState('selected')
          ) {
            graphRef.current?.setItemState(node, 'hover', true)
          }
          handleMouseEnter(evt)
        })

        graphRef.current?.on('node:mouseleave', evt => {
          handleMouseLeave(evt)
        })

        if (typeof window !== 'undefined') {
          window.onresize = () => {
            if (!graphRef.current || graphRef.current?.get('destroyed')) return
            if (
              !containerRef ||
              !containerRef.current?.scrollWidth ||
              !containerRef.current?.scrollHeight
            )
              return
            graphRef.current?.changeSize(
              containerRef?.current?.scrollWidth,
              containerRef.current?.scrollHeight,
            )
          }
        }
      } else {
        graphRef.current.clear()
        graphRef.current.changeData(topologyData)
        setTimeout(() => {
          graphRef.current.fitCenter()
        }, 100)
        setHightLight()
      }
    }
  }

  useImperativeHandle(drawRef, () => ({
    drawGraph,
  }))

  return (
    <div
      className={styles.g6_topology}
      style={{ height: isResource ? 450 : 400 }}
    >
      <div ref={containerRef} className={styles.g6_overview}>
        <div
          className={styles.g6_loading}
          style={{ display: topologyLoading ? 'block' : 'none' }}
        >
          <Loading />
        </div>
      </div>
    </div>
  )
})

export default TopologyMap
