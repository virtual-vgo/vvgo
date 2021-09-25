import React from 'react'
import ReactDOM from 'react-dom'
import {latestProject, useProjects} from "./models"
import {YoutubeIframe} from "./components"

export const Render = (selectors) => {
    const domContainer = document.querySelector(selectors)
    ReactDOM.render(<Index/>, domContainer)
}

const Index = (props) => {
    const [projects,] = useProjects()

    const latest = latestProject(projects)
    return <div className="mt-2 container">
        <div className="row">
            <div className="col col-12 col-lg-7">
                <Banner latest={latest}/>
                <YoutubeIframe latest={latest}/>
            </div>

            <div className={'col mt-2'}>
                <h3>Latest Projects</h3>
                <LatestProjects projects={projects}/>
                <h3>Member Highlights</h3>
                <MemberHighlights/>
            </div>
        </div>
        <div className="row justify-content-md-center text-center m-2">
            <div className="col text-center mt-2">
                <p>
                    If you would like to join our orchestra or get more information about our current projects,
                    please join us on <a href="https://discord.gg/9RVUJMQ">Discord!</a>
                </p>
            </div>
        </div>
    </div>
}

const Banner = (props) => {
    if (props.drawBanner === false) return <div/>
    if (props.latest === undefined) return <div/>

    const latest = props.latest
    const youtubeLink = latest.youtubeLink
    const bannerLink = latest.bannerLink

    const Banner = () => {
        if (bannerLink === "") return <div>
            <h1 className="title">{latest.title}</h1>
            <h2>{latest.sources}</h2>
        </div>
        else return <img src={bannerLink} className="mx-auto img-fluid" alt="banner"/>
    }

    return <div id='banner' className={'col'}>
        <a href={youtubeLink} className="text-light text-center">
            <Banner/>
        </a>
    </div>
}

const LatestProjects = (props) => {
    const projects = props.projects.filter(project => project.isOpenForSubmission())

    const Row = (props) => {
        const project = props.project
        return <tr>
            <td>
                <a href={project.partsPage} className="text-light">
                    {project.title} <br/> {project.sources}
                </a>
            </td>
        </tr>
    }

    return <table className="table text-light clickable">
        {projects.map(project => <Row key={project.name} project={project}/>)}
    </table>
}

const memberHighlightSrcs = [
    "https://cdn.discordapp.com/attachments/869388540272861245/869389373693640714/11-GS-Thomas.png",
    "https://cdn.discordapp.com/attachments/869388540272861245/870052230802313316/11-GS-Will.png",
    "https://cdn.discordapp.com/attachments/869388540272861245/870843556800135210/11-GS-Jordy.png",
    "https://cdn.discordapp.com/attachments/869388540272861245/871453949427855400/Artboard_1.png"
]
const memberHighlightSrc = memberHighlightSrcs[Math.floor(Math.random() * memberHighlightSrcs.length)]

const MemberHighlights = (props) => {
    return <table className="table text-light">
        <tr>
            <td>
                <img src={memberHighlightSrc} width="100%" alt="Thomas"/>
            </td>
        </tr>
    </table>
}
