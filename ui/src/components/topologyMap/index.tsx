import React, { forwardRef, useImperativeHandle, useRef, useState } from 'react'
import G6 from '@antv/g6'
import type {
  IG6GraphEvent,
  IGroup,
  ModelConfig,
  IAbstractGraph,
} from '@antv/g6'
import { useLocation } from 'react-router-dom'
import queryString from 'query-string'
import insertCss from 'insert-css';
import Loading from '@/components/loading'
import { ICON_MAP } from '@/utils/images'
import { capitalized } from '@/utils/tools'

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
  nodeData: {
    resourceType: string
    status: string
    resourceURN: string
    iamResourceID: string | number
    cloudResourceID: string | number
  },
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
  console.log(cfg.label, "sdasdsad")
  const list = cfg?.label?.split('.')
  const len = list?.length || 0
  return list?.[len - 1] || ''
}

type IProps = {
  topologyLoading?: boolean
  onTopologyNodeClick?: (node: any) => void
}

interface OverviewTooltipProps {
  type: string
  itemWidth: number
  hiddenButtonInfo: {
    x: number
    y: number
    e?: IG6GraphEvent
  }
  open: boolean
}

const statusColorMap = {
  destroyed: '#8a8a8a',
  applied: '#1778ff',
  failed: '#F5222D',
  unknown: '#F5222D',
}

const OverviewTooltip: React.FC<OverviewTooltipProps> = ({
  type,
  hiddenButtonInfo,
}) => {
  const model = hiddenButtonInfo?.e?.item?.get('model') as NodeModel
  const { nodeData } = model

  const boxStyle: any = {
    background: '#fff',
    border: '1px solid #f5f5f5',
    position: 'absolute',
    top: hiddenButtonInfo?.y || -500,
    left: hiddenButtonInfo?.x + 14 || -500,
    transform: 'translate(-50%, -100%)',
    zIndex: 5,
    padding: '6px 12px',
    borderRadius: 8,
    boxShadow: '0 2px 8px rgba(0,0,0,0.15)',
    borderTop: '2px solid #1677ff',
  }

  const itemStyle = {
    color: '#333',
    fontSize: 14,
    whiteSpace: 'nowrap',
  }

  const labelStyle = {
    color: '#999',
    fontSize: 14,
    marginRight: 5,
  }

  const statusStyleMap = {
    destroyed: {
      padding: '2px 5px',
      borderRadius: 6,
      color: 'rgba(0,0,0,0.88)',
      background: '#fafafa',
      border: '1px solid #d9d9d9'
    },
    applied: {
      border: '1px solid #91caff',
      padding: '2px 5px',
      borderRadius: 6,
      color: '#1778ff',
      background: '#e6f4ff',
    },
    failed: {
      padding: '2px 5px',
      borderRadius: 6,
      color: '#ff4d4f',
      background: '#fff2f0',
      border: '1px solid #ffccc7'
    }
  }

  return (
    <div style={boxStyle}>
      <div style={itemStyle}>
        {model?.label}
      </div>
      <div style={itemStyle}>
        <span style={labelStyle}>Type: </span>
        {nodeData?.resourceType}
      </div>
      <div style={itemStyle}>
        <span style={labelStyle}>Status: </span>
        <span style={statusStyleMap?.[nodeData?.status]}>
          {nodeData?.status}
        </span>
      </div>
      <div style={itemStyle}>
        <span style={labelStyle}>cloudResourceID: </span>
        {nodeData?.cloudResourceID}
      </div>
      <div style={itemStyle}>
        <span style={labelStyle}>iamResourceID: </span>
        {nodeData?.iamResourceID}
      </div>
      <div style={itemStyle}>
        <span style={labelStyle}>resourceURN: </span>
        {nodeData?.resourceURN}
      </div>
    </div>
  )
}

insertCss(`
  .g6-component-tooltip {
    // border-top: 2px solid #1677ff;
    // background-color: rgba(255, 255, 255, 0.8);
    background: #fff;
    box-shadow: rgb(174, 174, 174) 0px 0px 10px;
  }
  .tooltip-item {
    color: #333;
    font-size: 14px;
    white-space: nowrap;
  }
  .tooltip-item-label {
    color: #999;
    font-size: 14px;
    margin-right: 5px;
  }
  .tooltip-item-status {
    padding: 2px 5px;
    border-radius: 6px;
  }
  .tooltip-item-destroyed {
    color: rgba(0,0,0,0.88);
    background: #fafafa;
    border: 1px solid #d9d9d9;
  }
  .tooltip-item-applied {
    color: #1778ff;
    background: #e6f4ff;
    border: 1px solid #91caff;
  }
  .tooltip-item-failed,.tooltip-item-unknown {
    color: #ff4d4f;
    background: #fff2f0;
    border: 1px solid #ffccc7;
  }
`);


const tooltip = new G6.Tooltip({
  offsetX: 10,
  offsetY: 10,
  itemTypes: ['node', 'edge'],
  getContent: (e) => {
    const model = e?.item?.get('model') as NodeModel
    const { nodeData } = model as any
    const outDiv = document.createElement('div');
    outDiv.style.width = 'fit-content';
    //outDiv.style.padding = '0px 0px 20px 0px';
    outDiv.innerHTML = `
      <div class="tooltip-box">
        <div class="tooltip-item">
          ${model?.label}
        </div>
        <div class="tooltip-item">
          <span class="tooltip-item-label">Type: </span>
          ${nodeData?.resourceType}
        </div>
        <div class="tooltip-item">
          <span class="tooltip-item-label">Status: </span>
          <span class="tooltip-item-status tooltip-item-${nodeData?.status}">${nodeData?.status}</span>
        </div>
        <div class="tooltip-item">
          <span class="tooltip-item-label">cloudResourceID: </span>
          ${nodeData?.cloudResourceID}
        </div>
        <div class="tooltip-item">
          <span class="tooltip-item-label">iamResourceID: </span>
          ${nodeData?.iamResourceID}
        </div>
        <div class="tooltip-item">
          <span class="tooltip-item-label">resourceURN: </span>
          ${nodeData?.resourceURN}
        </div>
      </div>
    `;
    return outDiv;
  },
});

const TopologyMap = forwardRef((props: IProps, drawRef) => {
  const {
    onTopologyNodeClick,
    topologyLoading,
  } = props
  const containerRef = useRef(null)
  const graphRef = useRef<IAbstractGraph | null>(null)
  const location = useLocation()
  const { type } = queryString.parse(location?.search)
  const [tooltipopen, setTooltipopen] = useState(false)
  const [itemWidth, setItemWidth] = useState<number>(100)
  const [hiddenButtontooltip, setHiddenButtontooltip] = useState<{
    x: number
    y: number
    e?: IG6GraphEvent
  }>({ x: -500, y: -500, e: undefined })


  function handleMouseEnter(evt) {
    graphRef.current?.setItemState(evt.item, 'hoverState', true)
    const bbox = evt.item.getBBox()
    const point = graphRef.current?.getCanvasByPoint(bbox.centerX, bbox.minY)
    if (bbox) {
      setItemWidth(bbox.width)
    }
    setHiddenButtontooltip({ x: point.x, y: point.y - 5, e: evt })
    setTooltipopen(true)
  }

  const handleMouseLeave = (evt: IG6GraphEvent) => {
    graphRef.current?.setItemState(evt.item, 'hoverState', false)
    setTooltipopen(false)
  }

  G6.registerNode(
    'card-node',
    {
      draw(cfg: NodeConfig, group: IGroup) {
        const displayName = getNodeName(cfg, type as string)
        const count = cfg.data?.count
        const nodeWidth = 235
        const nodeHeight = 55

        // Create main container
        const rect = group.addShape('rect', {
          attrs: {
            x: 0,
            y: 0,
            width: nodeWidth,
            height: nodeHeight,
            radius: 6,
            fill: '#ffffff',
            stroke: '#e6f4ff',
            // fill: statusColorMap?.[(cfg?.nodeData as any)?.status],
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
            height: nodeHeight,
            radius: [3, 0, 0, 3],
            fill: statusColorMap?.[(cfg?.nodeData as any)?.status],
            opacity: 0.4,
          },
          name: 'node-accent',
        })

        // Add Kubernetes icon
        const iconSize = 40
        const resourcePlane = (cfg?.nodeData as any)?.resourcePlane
        const typeList = (cfg?.nodeData as any)?.resourceType?.split('/')
        const nodeType = resourcePlane === 'Kubernetes' ? typeList?.[typeList?.length - 1] : resourcePlane
        const iconHeight = resourcePlane === 'aws' ? 30 : resourcePlane === 'azure' ? 27 : iconSize
        group.addShape('image', {
          attrs: {
            x: 5,
            y: (nodeHeight - iconHeight) / 2,
            width: iconSize,
            height: iconHeight,
            img: ICON_MAP?.[nodeType as keyof typeof ICON_MAP] || ICON_MAP.Kubernetes,
          },
          name: 'node-icon',
        })

        // Add title text
        group.addShape('text', {
          attrs: {
            x: 52,
            y: 25,
            text: fittingString(displayName || '', 140, 14),
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

        const statusShape0 = group.addShape('circle', {
          attrs: {
            zIndex: -4,
            x: 220,
            y: 20,
            fill: statusColorMap?.[(cfg?.nodeData as any)?.status],
            lineWidth: 1,
            r: 3,
            opacity: 0.4,
          },
          name: 'status-circle1',
        })

        const statusShape1 = group.addShape('circle', {
          attrs: {
            zIndex: -3,
            x: 220,
            y: 20,
            fill: statusColorMap?.[(cfg?.nodeData as any)?.status],
            r: 3,
            opacity: 0.4,
            lineWidth: 1,
          },
          name: 'status-circle1',
        })
        const statusShape2 = group.addShape('circle', {
          attrs: {
            zIndex: -2,
            x: 220,
            y: 20,
            fill: statusColorMap?.[(cfg?.nodeData as any)?.status],
            r: 3,
            opacity: 0.4,
            lineWidth: 1,
          },
          name: 'status-circle2',
        })
        const statusShape3 = group.addShape('circle', {
          attrs: {
            zIndex: -1,
            x: 220,
            y: 20,
            fill: statusColorMap?.[(cfg?.nodeData as any)?.status],
            r: 3,
            opacity: 0.4,
            lineWidth: 1,
          },
          name: 'status-circle3',
        })
        group.sort(); // Sort according to the zIndex

        statusShape0.animate({
          r: 8,
          opacity: 0.1,
        }, {
          duration: 4000,
          easing: 'easeLinear',
          delay: 0,
          repeat: true,
        },)
        statusShape1.animate({
          r: 8,
          opacity: 0.1,
        }, {
          duration: 4000,
          easing: 'easeLinear',
          delay: 1000,
          repeat: true,
        },)
        statusShape2.animate({
          r: 8,
          opacity: 0.1,
        }, {
          duration: 4000,
          easing: 'easeLinear',
          delay: 2000,
          repeat: true,
        },)
        statusShape3.animate({
          r: 8,
          opacity: 0.1,
        }, {
          duration: 4000,
          easing: 'easeLinear',
          delay: 3000,
          repeat: true,
        },)

        group.addShape('text', {
          attrs: {
            x: 52,
            y: 40,
            text: capitalized(nodeType),
            fontSize: 10,
            fontWeight: 400,
            fill: '#8a8a8a',
            cursor: 'pointer',
            textBaseline: 'middle',
            fontFamily:
              '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial',
          },
          name: 'node-type',
        })

        return rect
      },
      setState(name, value, item) {
        const shape = item.get('keyShape')
        const status = (item.getModel()?.nodeData as any)?.status
        if (name === 'hover' || name === 'hoverState') {
          shape?.attr('stroke', value ? statusColorMap?.[status] : '#fff')
          shape?.attr('lineWidth', 1)
          shape?.attr('strokeOpacity', value ? 0.7 : 1)
        }
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
      plugins: [toolbar, tooltip],
      enabledStack: true,
      modes: {
        default: ['drag-canvas', 'drag-node', 'click-select'],
      },
      animate: true,
      layout: {
        type: 'dagre',
        rankdir: 'LR',
        align: 'UL',
        nodesep: 25,
        ranksep: 30,
        nodesepFunc: () => 1,
        ranksepFunc: () => 1,
        controlPoints: true,
        sortByCombo: false,
        preventOverlap: true,
        nodeSize: [200, 60],
        workerEnabled: true,
        clustering: false,
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
        size: [235, 55],
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

  function drawGraph(topologyData) {
    if (topologyData) {
      if (graphRef.current) {
        graphRef.current?.destroy()
        graphRef.current = null
      }
      if (!graphRef.current) {
        graphRef.current = initGraph()
        graphRef.current?.read(topologyData)

        setTimeout(() => {
          graphRef.current.fitCenter()
        }, 100)

        graphRef.current?.on('node:click', evt => {
          const node = evt.item
          const model = node.getModel()
          setTooltipopen(false)
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
      }
    }
  }

  useImperativeHandle(drawRef, () => ({
    drawGraph,
  }))

  return (
    <div
      className={styles.g6_topology}
      style={{ height: 800 }}
    >
      <div className={styles.g6_node_status}>
        {
          Object.entries(statusColorMap)?.map(([key, value]) => {
            return <div key={key} className={styles.status_item}>
              {key}<span className={styles.status_dot} style={{ background: value }}></span>
            </div>
          })
        }
      </div>
      <div ref={containerRef} className={styles.g6_overview}>
        <div
          className={styles.g6_loading}
          style={{ display: topologyLoading ? 'block' : 'none' }}
        >
          <Loading />
        </div>
        {/* {tooltipopen ? (
          <OverviewTooltip
            type={type as string}
            itemWidth={itemWidth}
            hiddenButtonInfo={hiddenButtontooltip}
            open={tooltipopen}
          />
        ) : null} */}
      </div>
    </div>
  )
})

export default TopologyMap
