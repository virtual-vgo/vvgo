import {hot} from 'react-hot-loader'
import React, {useState} from 'react'
import ReactDOM from 'react-dom'
import {useCredits, useProjects} from "./models";

const BannerFadeoutStart = 3000
const BannerFadeoutEnd = BannerFadeoutStart + 2000

export const Render = () => {
    const domContainer = document.querySelector('#reactDom')
    ReactDOM.render(<Feature/>, domContainer)
}

const Feature = (props) => {
    const [projects,] = useProjects()

    const [showBanner, setShowBanner] = useState(true)
    new Promise((resolve) => setTimeout(resolve, BannerFadeoutStart)).then(_ => setShowBanner(false))

    const [drawBanner, setDrawBanner] = useState(true)
    new Promise((resolve) => setTimeout(resolve, BannerFadeoutEnd)).then(_ => setDrawBanner(false))

    const [showCredits, setShowCredits] = useState(false)
    const toggleCredits = () => setShowCredits(!showCredits)

    const latest = latestProject(projects)
    const [credits,] = useCredits(latest)

    if (latest === undefined) return <div/>
    return <div className='container'>
        <div className='row row-cols-1'>
            <Banner latest={latest} showBanner={showBanner} drawBanner={drawBanner}/>
            <Video latest={latest} drawBanner={drawBanner}/>
            <Credits latest={latest} credits={credits} showCredits={showCredits} toggleCredits={toggleCredits}/>
        </div>
    </div>
}

const Banner = (props) => {
    const style = (props.showBanner) ? 'visible' : 'hidden'
    const latest = props.latest
    if (latest === undefined) return <div/>
    const youtubeLink = latest.youtubeLink
    const bannerLink = latest.bannerLink
    if (props.drawBanner === false) return <div/>
    return <div id='banner' className={['col', style].join(' ')}>
        <a href={youtubeLink} className="btn btn-link nav-link">
            <img src={bannerLink} className="mx-auto img-fluid" alt="banner"/>
        </a>
    </div>
}

const Video = (props) => {
    if (props.drawBanner) return <div/>

    const latest = props.latest
    if (latest === undefined) return <div/>
    else if (latest.youtubeEmbed === null) return <div/>
    else if (latest.youtubeEmbed.startsWith('https://') === false) return <div/>
    else return <div className='col'>
            <div className='project-iframe-wrapper text-center m-2'>
                <iframe className='project-iframe' src={latest.youtubeEmbed}
                        allow='accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture'
                        allowFullScreen/>
            </div>
            <div className="text-right font small text-uppercase">
                {latest.composers} <br/>
                {latest.arrangers}
            </div>
        </div>
}

const Credits = (props) => {
    if (props.showCredits) {
        const credits = props.credits
        console.log(props.credits)

        return <div>
            <div className={'btn btn-link'} onClick={(_) => props.toggleCredits()}>
                <p className={'text-left'}>Hide credits.</p>
            </div>
            {credits.map(topicRow => <CreditsTopic key={topicRow.name} topicRow={topicRow}/>)}
        </div>
    } else {
        return <div className={'btn btn-link'} onClick={(_) => props.toggleCredits()}>
            <p className={'text-left text-center text-light'}>Show credits.</p>
        </div>
    }
}

const CreditsTopic = (props) => {
    const topicRow = props.topicRow
    return <div>
        <div className="row">
            <div className="col text-center">
                <h2><strong>-- {topicRow.name} --</strong></h2>
            </div>
        </div>
        <div className="card-columns">
            {topicRow.rows.map(credit => <CreditsTeam credit={credit}/>)}
        </div>
    </div>
}

const CreditsTeam = (props) => {
    const credit = props.credit
    return <div className="card bg-transparent text-center">
        <h5>{credit.name}</h5>
        <ul className="list-unstyled">
            {credit.rows.map(x => <li>{x.name} <small>{x.bottomText}</small></li>)}
        </ul>
    </div>
}

const latestProject = (projects) => {
    console.log(projects)
    const released = projects.filter(proj => proj.videoReleased === true)
    released.sort()
    return released.pop()
}

export default hot(module)(Feature)
