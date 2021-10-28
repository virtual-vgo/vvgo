import Col from "react-bootstrap/Col";
import Row from "react-bootstrap/Row";
import {Link} from "react-router-dom";
import {Highlight, latestProject, Project, projectIsOpenForSubmission, useHighlights, useProjects} from "../datasets";
import {randElement} from "../utils";
import {RootContainer} from "./shared/RootContainer";
import {YoutubeIframe} from "./shared/YoutubeIframe";
import React = require("react");

export const Home = () => {
    const highlights = useHighlights();
    const projects = useProjects();
    const highlight = randElement(highlights);

    const latest = latestProject(projects);
    return <RootContainer>
        <Row>
            <Col lg={7} md={12}>
                <Banner latest={latest}/>
                <YoutubeIframe latest={latest}/>
            </Col>
            <Col>
                <div className={"col mt-2"}>
                    <h3>Latest Projects</h3>
                    <LatestProjects projects={projects}/>
                    <h3>Member Highlights</h3>
                    <MemberHighlight highlight={highlight}/>
                </div>
                <div className="row justify-content-md-center text-center m-2">
                    <div className="col text-center mt-2">
                        <p>
                            If you would like to join our orchestra or get more information about our current projects,
                            please join us on <a href="https://discord.gg/9RVUJMQ">Discord!</a>
                        </p>
                    </div>
                </div>
            </Col>
        </Row>
    </RootContainer>;
};

const Banner = (props: { latest: Project }) => {
    if (!props.latest) return <div/>;

    const latest = props.latest;
    const youtubeLink = latest.YoutubeLink;
    const bannerLink = latest.BannerLink;

    const Banner = () => {
        if (bannerLink === "") return <div>
            <h1 className="title">{latest.Title}</h1>
            <h2>{latest.Sources}</h2>
        </div>;
        else return <img src={bannerLink} className="mx-auto img-fluid" alt="banner"/>;
    };

    return <div id="banner" className={"col"}>
        <a href={youtubeLink} className="text-light text-center">
            <Banner/>
        </a>
    </div>;
};

const LatestProjects = (props: { projects: Project[] }) => {
    if (!props.projects) return <div/>;
    const projects = props.projects.filter((project: Project): boolean => {
        return projectIsOpenForSubmission(project);
    });

    const Row = (props: { project: Project }) => {
        const project = props.project;
        return <tr>
            <td>
                <Link to="/parts/" className="text-light">
                    {project.Title} <br/> {project.Sources}
                </Link>
            </td>
        </tr>;
    };

    return <table className="table text-light clickable">
        <tbody>
        {projects.map(project => <Row key={project.Name} project={project}/>)}
        </tbody>
    </table>;
};

const MemberHighlight = (props: { highlight: Highlight }) => {
    if (!props.highlight) return <div/>;
    const {Source, Alt} = props.highlight;
    return <table className="table text-light">
        <tbody>
        <tr>
            <td>
                <img src={Source} width="100%" alt={Alt}/>
            </td>
        </tr>
        </tbody>
    </table>;
};
