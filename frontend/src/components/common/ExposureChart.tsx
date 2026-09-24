import ReactECharts from 'echarts-for-react'
import type { CompartmentCurve } from '@/types/assessment'
import type { ExposureSegment } from '@/types/segment'

interface Props { segments: ExposureSegment[]; curves?: CompartmentCurve[]; height?: number }

export function ExposureChart({ segments, curves = [], height = 360 }: Props) {
  let elapsed = 0
  const depth = [[0, 0] as [number, number], ...segments.map((segment) => {
    elapsed += segment.duration_min
    return [elapsed, segment.depth_m] as [number, number]
  })]
  const series = [
    { name: 'Depth input', type: 'line', yAxisIndex: 0, data: depth, step: 'end', symbolSize: 7, lineStyle: { width: 3, color: '#176c64' }, itemStyle: { color: '#176c64' }, areaStyle: { color: 'rgba(23,108,100,.10)' } },
    ...curves.slice(0, 6).map((curve, index) => ({
      name: `${curve.name} inert load`, type: 'line', yAxisIndex: 1,
      data: curve.points.map((point) => [point.elapsed_min, point.total_inert_bar]),
      showSymbol: false, lineStyle: { width: index < 2 ? 2.2 : 1.2, opacity: index < 2 ? 1 : .65 },
    })),
  ]
  const option = {
    animationDuration: 360,
    color: ['#176c64', '#b04a37', '#d39b28', '#3f6c99', '#806397', '#64736d', '#9a7452'],
    grid: { left: 58, right: 58, top: 54, bottom: 52 },
    legend: { top: 8, left: 0, textStyle: { color: '#44514e', fontSize: 11 } },
    tooltip: { trigger: 'axis', backgroundColor: '#173b37', borderWidth: 0, textStyle: { color: '#f3f6f2' } },
    xAxis: { type: 'value', name: 'Elapsed min', nameLocation: 'middle', nameGap: 32, axisLine: { lineStyle: { color: '#98a6a2' } }, splitLine: { lineStyle: { color: '#e4e9e6' } } },
    yAxis: [
      { type: 'value', name: 'Depth m', inverse: true, min: 0, axisLine: { show: true, lineStyle: { color: '#98a6a2' } }, splitLine: { lineStyle: { color: '#e4e9e6' } } },
      { type: 'value', name: 'Inert load bar', position: 'right', axisLine: { show: true, lineStyle: { color: '#b7a27c' } }, splitLine: { show: false } },
    ],
    series,
  }
  return <div className="chart-shell"><ReactECharts option={option} style={{ height }} notMerge /></div>
}
