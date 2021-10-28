import React = require("react");
import Button from "react-bootstrap/Button";
import ButtonGroup from "react-bootstrap/ButtonGroup";
import Col from "react-bootstrap/Col";
import Row from "react-bootstrap/Row";
import Table from "react-bootstrap/Table";
import {Part, Project, useParts, useProjects} from "../datasets";
import {LoadingText} from "./shared/LoadingText";
import {RootContainer} from "./shared/RootContainer";

const recordingInstructions = "https://cdn.discordapp.com/attachments/741188776088436748/799697926661210212/VVGO_RecordingInstructions_Season2.png";

export const Parts = () => {
    const allProjects = useProjects()
        .filter(r => r.Name && r.Name.length > 0)
        .filter(r => r.Title && r.Title.length > 0)
        .sort((a, b) => -a.Name.localeCompare(b.Name));
    const parts = useParts();

    const [showProject, setShowProject] = React.useState("");

    if (showProject == "" && allProjects.length > 0)
        setShowProject(allProjects[0].Name);

    const project = allProjects.filter(r => r.Name == showProject).pop();

    if (allProjects.length == 0)
        return <RootContainer>
            <LoadingText/>
        </RootContainer>;

    return <RootContainer>
        <Row>
            <Col lg={3}>
                <ButtonGroup vertical>
                    {allProjects.map(project =>
                        <Button
                            variant="outline-light"
                            className={"bg-transparent text-light"}
                            key={project.Name}
                            onClick={() => setShowProject(project.Name)}>
                            {project.Title}
                        </Button>)}
                </ButtonGroup>
            </Col>
            <Col>
                <ProjectInfo project={project}/>
                <PartsTopLinks project={project}/>
                <PartsTable projectName={showProject} parts={parts}/>
            </Col>
        </Row>
    </RootContainer>;
};

const ProjectInfo = (props: { project: Project }) =>
    props.project ? <div>
        <h1>{props.project.Title}</h1>
    </div> : <div/>;

export const PartsTopLinks = (props: { project: Project }) => {
    const Card = (props: {
        to: string,
        children: (string | JSX.Element)[]
    }) => <Button
        variant="outline-light"
        className="btn-lnk"
        href={props.to}>
        {props.children}</Button>;

    return <ButtonGroup>
        <Card to={recordingInstructions}>
            <i className="far fa-image"/> Recording Instructions
        </Card>
        <Card to={props.project.ReferenceTrackLink}>
            <i className="far fa-file-audio"/> Reference Track
        </Card>
        <Card to={props.project.SubmissionLink}>
            <i className="fab fa-dropbox"/> Submit Recordings
        </Card>
    </ButtonGroup>;
};

const PartsTable = (props: { projectName: string, parts: Part[] }) =>
    <Table className="text-light">
        <thead>
        <tr>
            <th>Part</th>
            <th>Downloads</th>
        </tr>
        </thead>
        <tbody>
        {props.parts.filter(part => props.projectName == part.Project)
            .map(part => <tr key={part.PartName}>
                <td>{part.PartName}</td>
                <td><PartDownloads part={part}/></td>
            </tr>)}
        </tbody>
    </Table>;

const PartDownloads = (props: { part: Part }) => {
    const Link = (props: {
        to: string,
        children: string | (string | JSX.Element)[]
    }) => props.to && props.to.length > 0 ?
        <Col className="col-sm-auto text-nowrap">
            <Button href={props.to} className="btn-sm btn-link btn-outline-light bg-dark text-light">
                {props.children}
            </Button>
        </Col> : <div/>;

    return <Row className="justify-content-start">
        <Link to={props.part.SheetMusicLink}><i className="far fa-file-pdf"/> sheet music</Link>
        <Link to={props.part.ClickTrackLink}><i className="far fa-file-audio"/> click track</Link>
        <Link to={props.part.ConductorVideo}><i className="far fa-file-video"/> conductor video</Link>
        <Link to={props.part.PronunciationGuideLink}><i className="fas fa-language"/> pronunciation guide</Link>
    </Row>;
};
