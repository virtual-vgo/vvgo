import React from 'react'
import {latestProject, useProjects} from "./models"
import {YoutubeIframe} from "./components"
import {Render} from "./render";

export const Entrypoint = (selectors) => Render(<Index/>, selectors)

const Index = (props) => {
    const [projects, _] = useProjects()

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
    if (props.latest === undefined) return <div/>

    const latest = props.latest
    const youtubeLink = latest.YoutubeLink
    const bannerLink = latest.BannerLink

    const Banner = () => {
        if (bannerLink === "") return <div>
            <h1 className="title">{latest.Title}</h1>
            <h2>{latest.Sources}</h2>
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
                <a href={project.PartsPage} className="text-light">
                    {project.Title} <br/> {project.Sources}
                </a>
            </td>
        </tr>
    }

    return <table className="table text-light clickable">
        <tbody>
        {projects.map(project => <Row key={project.Name} project={project}/>)}
        </tbody>
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
        <tbody>
        <tr>
            <td>
                <img src={memberHighlightSrc} width="100%" alt="Thomas"/>
            </td>
        </tr>
        </tbody>
    </table>
}
