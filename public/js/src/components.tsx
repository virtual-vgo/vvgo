import {useEffect, useRef} from 'react'
import {Project} from "./datasets";
import React = require('react');
import d3 = require('d3');

export const Container = (props: JSX.IntrinsicAttributes & React.ClassAttributes<HTMLDivElement> & React.HTMLAttributes<HTMLDivElement>) =>
    <div className={'mt-4 container'} {...props}/>

export const Visualization = (props: { drawSVG: (arg0: d3.Selection<SVGSVGElement, unknown, HTMLElement, any>, arg1: any) => void; }) => {
    const ref = useRef()
    const div = <div ref={ref}/>
    const svg = d3.select(ref.current).append("svg")
    useEffect((): any => {
        props.drawSVG(svg, props)
        return () => svg.remove()
    })
    return div
}

export const YoutubeIframe = (props: { latest: Project }) => {
    const latest = props.latest
    if (latest === undefined) return <div/>
    if (latest.YoutubeEmbed === "") return <div/>
    return <div className='project-iframe-wrapper text-center m-2'>
        <iframe className='project-iframe' src={latest.YoutubeEmbed}
                allow='accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture'
                allowFullScreen/>
    </div>
}
