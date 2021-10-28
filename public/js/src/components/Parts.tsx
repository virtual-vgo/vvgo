import React = require("react");
import Button from "react-bootstrap/Button";
import ButtonGroup from "react-bootstrap/ButtonGroup";
import Col from "react-bootstrap/Col";
import Row from "react-bootstrap/Row";
import Table from "react-bootstrap/Table";
import {getSession} from "../auth";
import {Channels} from "../data/discord";
import {RecordingInstructions} from "../data/downloadLinks";
import {Part, Project, projectIsOpenForSubmission, useParts, useProjects, UserRoles} from "../datasets";
import {LinkChannel} from "./shared/LinkChannel";
import {LoadingText} from "./shared/LoadingText";
import {ProjectBanner} from "./shared/ProjectBanner";
import {RootContainer} from "./shared/RootContainer";
import {ShowHideToggle} from "./shared/ShowHideToggle";

export const Parts = () => {
    const me = getSession();
    const allProjects = useProjects();
    const parts = useParts();

    const [project, setProject] = React.useState(null as Project);
    const [showUnreleased, setShowUnreleased] = React.useState(false);
    const [showArchived, setShowArchived] = React.useState(false);

    if (!(allProjects && parts))
        return <RootContainer><LoadingText/></RootContainer>;

    const wantProjects = allProjects
        .filter(r => showUnreleased || r.PartsReleased == true)
        .filter(r => showArchived || r.PartsArchived == false);

    if (project == null && wantProjects.length > 0) setProject(wantProjects[0]);
    return <RootContainer>
        <Row>
            <Col lg={3}>
                <Row>
                    <Col>
                        {me.Roles.includes(UserRoles.ProductionTeam) ?
                            <ShowHideToggle
                                title="Unreleased"
                                state={showUnreleased}
                                setState={setShowUnreleased}/> : ""}
                    </Col>
                    <Col>
                        {me.Roles.includes(UserRoles.ExecutiveDirector) ?
                            <ShowHideToggle
                                title="Archived"
                                state={showArchived}
                                setState={setShowArchived}/> : ""}
                    </Col>
                </Row>
                <ButtonGroup vertical className="m-2">
                    {wantProjects.map(want =>
                        <Button
                            variant={projectIsOpenForSubmission(want) ? "outline-light" : "outline-warning"}
                            key={want.Name}
                            onClick={() => setProject(want)}>
                            {want.Title}
                            {want.PartsReleased == false ? <em><small><br/>Unreleased</small></em> : ""}
                            {want.PartsArchived == true ? <em><small><br/>Archived</small></em> : ""}
                        </Button>)}
                </ButtonGroup>
            </Col>
            {project ?
                <Col>
                    <ProjectHeader project={project}/>
                    <PartsTopLinks project={project}/>
                    <PartsTable projectName={project.Name} parts={parts}/>
                </Col> :
                <Col>
                    <p>There are no projects currently accepting submissions, but we are working hard to bring you some!
                        <br/>Please check <LinkChannel channel={Channels.NextProjectHints}/> for updates.</p>
                </Col>}
        </Row>
    </RootContainer>;
};

const ProjectHeader = (props: { project: Project }) => {
    return <Row className="row-cols-1">
        <Col>
            {props.project.PartsArchived ?
                <div className="alert alert-warning">
                    This project has been archived. Parts are only visible to executive directors.
                </div> : <div/>}
        </Col>
        <Col>
            {props.project.PartsReleased == false ?
                <div className="alert alert-warning">
                    This project is unreleased and invisible to members!
                </div> : <div/>}
        </Col>
        <Col className="text-center">
            <ProjectBanner project={props.project}/>
            <h3>{props.project.Sources}</h3>
            {props.project.Composers}
            <br/><small>{props.project.Arrangers}</small>
            <div className="m-2">
                <h4><strong>Submission Deadline:</strong>
                    <em>{props.project.SubmissionDeadline} (Hawaii Time)</em></h4>
            </div>
        </Col>
    </Row>;
};

const ButtonGroupBreakPoint = 800;

export const PartsTopLinks = (props: { project: Project }) => {
    const Card = (props: {
        to: string,
        children: (string | JSX.Element)[]
    }) => <Button
        variant="outline-light"
        className="btn-lnk"
        href={props.to}>
        {props.children}</Button>;

    return <ButtonGroup vertical={(window.visualViewport.width < ButtonGroupBreakPoint)}>
        <Card to={RecordingInstructions}>
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
        <Button href={props.to} className="btn-sm btn-link btn-outline-light bg-dark text-light">
            {props.children}
        </Button> :
        <div/>;

    return <ButtonGroup
        className="justify-content-start"
        vertical={(window.visualViewport.width < ButtonGroupBreakPoint)}>
        <Link to={props.part.SheetMusicLink}><i className="far fa-file-pdf"/> sheet music</Link>
        <Link to={props.part.ClickTrackLink}><i className="far fa-file-audio"/> click track</Link>
        <Link to={props.part.ConductorVideo}><i className="far fa-file-video"/> conductor video</Link>
        <Link to={props.part.PronunciationGuideLink}><i className="fas fa-language"/> pronunciation guide</Link>
    </ButtonGroup>;
};
