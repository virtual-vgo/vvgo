import React = require("react");
import * as _ from "lodash";
import Button from "react-bootstrap/Button";
import ButtonGroup from "react-bootstrap/ButtonGroup";
import Col from "react-bootstrap/Col";
import FormControl from "react-bootstrap/FormControl";
import Row from "react-bootstrap/Row";
import Table from "react-bootstrap/Table";
import {getSession} from "../auth";
import {Channels} from "../data/discord";
import {links} from "../data/links";
import {Part, Project, projectIsOpenForSubmission, useParts, useProjects, UserRoles} from "../datasets";
import {AlertArchivedParts} from "./shared/AlertArchivedParts";
import {AlertUnreleasedProject} from "./shared/AlertUnreleasedProject";
import {LinkChannel} from "./shared/LinkChannel";
import {LoadingText} from "./shared/LoadingText";
import {ProjectHeader} from "./shared/ProjectHeader";
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

    return <RootContainer title="Parts">
        <Row>
            <Col lg={3}>
                <div className={"d-flex flex-row justify-content-center"}>
                    {me.Roles.includes(UserRoles.ProductionTeam) ?
                        <ShowHideToggle
                            title="Unreleased"
                            state={showUnreleased}
                            setState={setShowUnreleased}/> : ""}

                    {me.Roles.includes(UserRoles.ExecutiveDirector) ?
                        <ShowHideToggle
                            title="Archived"
                            state={showArchived}
                            setState={setShowArchived}/> : ""}
                </div>
                <div className="d-flex justify-content-center">
                    <ButtonGroup vertical className="m-2">
                        {wantProjects.map(want =>
                            <Button
                                variant={project && project.Name == want.Name ?
                                    projectIsOpenForSubmission(want) ? "light" : "warning" :
                                    projectIsOpenForSubmission(want) ? "outline-light" : "outline-warning"
                                }
                                key={want.Name}
                                onClick={() => setProject(want)}>
                                {want.Title}
                                {want.PartsReleased == false ? <em><small><br/>Unreleased</small></em> : ""}
                                {want.PartsArchived == true ? <em><small><br/>Archived</small></em> : ""}
                            </Button>)}
                    </ButtonGroup>
                </div>
            </Col>
            {project ?
                <Col className="mx-4">
                    <AlertArchivedParts project={project}/>
                    <AlertUnreleasedProject project={project}/>
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

    return <div className="d-flex justify-content-center">
        <ButtonGroup vertical={(window.visualViewport.width < ButtonGroupBreakPoint)}>
            <Card to={links.RecordingInstructions}>
                <i className="far fa-image"/> Recording Instructions
            </Card>
            <Card to={props.project.ReferenceTrackLink}>
                <i className="far fa-file-audio"/> Reference Track
            </Card>
            <Card to={props.project.SubmissionLink}>
                <i className="fab fa-dropbox"/> Submit Recordings
            </Card>
        </ButtonGroup>
    </div>;
};

const PartsTable = (props: { projectName: string, parts: Part[] }) => {
    const [searchInput, setSearchInput] = React.useState("");
    const searchInputRef = React.useRef({} as HTMLInputElement);

    const wantParts = props.parts
        .filter(p => p.PartName.toLowerCase().includes(searchInput))
        .filter(p => p.Project == props.projectName);

    const searchBoxStyle = {maxWidth: 250} as React.CSSProperties;
    // This width gives enough space to have all the download buttons on one line
    const partNameStyle = {width: 220} as React.CSSProperties;

    return <div className="d-flex justify-content-center">
        <div className="d-flex flex-column flex-fill justify-content-center">
            <FormControl
                className="mt-4"
                style={searchBoxStyle}
                ref={searchInputRef}
                placeholder="Search Parts"
                onChange={() => setSearchInput(searchInputRef.current.value.toLowerCase())}/>
            <Table className="text-light">
                <thead>
                <tr>
                    <th>Part</th>
                    <th>Downloads</th>
                </tr>
                </thead>
                <tbody>
                {wantParts.map(part =>
                    <tr key={part.PartName}>
                        <td style={partNameStyle}>{part.PartName}</td>
                        <td><PartDownloads part={part}/></td>
                    </tr>)}
                </tbody>
            </Table>
        </div>
    </div>;
};

const DownloadButton = (props: {
    fileName: string,
    children: string | (string | JSX.Element)[]
}) => {
    const params = new URLSearchParams({fileName: props.fileName, token: getSession().Key});
    return <Button
        href={"/download?" + params.toString()}
        variant="outline-light"
        size={"sm"}>
        {props.children}
    </Button>;
};

const PartDownloads = (props: { part: Part }) => {
    const buttons = [] as Array<JSX.Element>;
    if (_.isEmpty(props.part.SheetMusicFile) == false)
        buttons.push(<DownloadButton
            key={props.part.SheetMusicFile}
            fileName={props.part.SheetMusicFile}>
            <i className="far fa-file-pdf"/> sheet music
        </DownloadButton>);

    if (_.isEmpty(props.part.ClickTrackFile) == false)
        buttons.push(<DownloadButton
            key={props.part.ClickTrackFile}
            fileName={props.part.ClickTrackFile}>
            <i className="far fa-file-audio"/> click track
        </DownloadButton>);

    if (_.isEmpty(props.part.ConductorVideo) == false)
        buttons.push(<Button
            key={props.part.ConductorVideo}
            href={props.part.ConductorVideo}
            variant="outline-light"
            size={"sm"}>
            <i className="far fa-file-video"/> conductor video
        </Button>);

    if (_.isEmpty(props.part.PronunciationGuide) == false)
        buttons.push(<DownloadButton
            key={props.part.PronunciationGuide}
            fileName={props.part.PronunciationGuide}>
            <i className="fas fa-language"/> pronunciation guide
        </DownloadButton>);

    return <ButtonGroup
        className="justify-content-start"
        vertical={(window.visualViewport.width < ButtonGroupBreakPoint)}>
        {buttons}
    </ButtonGroup>;
};
