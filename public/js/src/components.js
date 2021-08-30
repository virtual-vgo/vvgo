import React from 'react'

export const YoutubeIframe = (props) => {
    const latest = props.latest
    if (latest === undefined) return <div/>
    if (latest.youtubeEmbed === "") return <div/>
    return <div className='project-iframe-wrapper text-center m-2'>
        <iframe className='project-iframe' src={latest.youtubeEmbed}
                allow='accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture'
                allowFullScreen/>
    </div>
}
