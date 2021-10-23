import React = require('react');
import {Highlight, latestProject, Project, projectIsOpenForSubmission, useHighlights, useProjects} from "./datasets"
import {Container, YoutubeIframe} from "./components"
import {Render} from "./render";
import {randElement} from "./utils";

export const Entrypoint = (selectors: string) => Render(<Index/>, selectors)

const Index = () => {
    const highlights = useHighlights()
    const projects = useProjects()
    const highlight = highlights && highlights.length > 0 ? randElement(highlights) : new Highlight()

    const latest = latestProject(projects)
    return <Container>
        <div className="row">
            <div className="col col-12 col-lg-7">
                <Banner latest={latest}/>
                <YoutubeIframe latest={latest}/>
            </div>

            <div className={'col mt-2'}>
                <h3>Latest Projects</h3>
                <LatestProjects projects={projects}/>
                <h3>Member Highlights</h3>
                <MemberHighlight highlight={highlight}/>
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
    </Container>
}

const Banner = (props: { latest: Project }) => {
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

const LatestProjects = (props: { projects: Project[] }) => {
    const projects = props.projects.filter((project: Project): boolean => {
        return projectIsOpenForSubmission(project)
    })

    const Row = (props: { project: Project }) => {
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

const MemberHighlight = (props: { highlight: Highlight }) => {
    return <table className="table text-light">
        <tbody>
        <tr>
            <td>
                <img src={props.highlight.Source} width="100%" alt={props.highlight.Alt}/>
            </td>
        </tr>
        </tbody>
    </table>
}
