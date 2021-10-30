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
import {
    ApiRole,
    Part,
    Project,
    projectIsOpenForSubmission,
    Session,
    useNewApiSession,
    useParts,
    useProjects,
    UserRole,
} from "../datasets";
import {AlertArchivedParts} from "./shared/AlertArchivedParts";
import {AlertUnreleasedProject} from "./shared/AlertUnreleasedProject";
import {LinkChannel} from "./shared/LinkChannel";
import {LoadingText} from "./shared/LoadingText";
import {ProjectHeader} from "./shared/ProjectHeader";
import {RootContainer} from "./shared/RootContainer";
import {ShowHideToggle} from "./shared/ShowHideToggle";

export const Parts = () => {
    const documentTitle = "Parts";
    const me = getSession();
    const allProjects = useProjects();
    const parts = useParts();
    const downloadSession = useNewApiSession(4 * 3600, [ApiRole.Download]);

    const [project, setProject] = React.useState(null as Project);
    const [showUnreleased, setShowUnreleased] = React.useState(false);
    const [showArchived, setShowArchived] = React.useState(false);

    if (!(allProjects && parts))
        return <RootContainer title={documentTitle}><LoadingText/></RootContainer>;

    const wantProjects = allProjects
        .filter(r => showUnreleased || r.PartsReleased == true)
        .filter(r => showArchived || r.PartsArchived == false);

    if (project == null && wantProjects.length > 0) setProject(wantProjects[0]);

    return <RootContainer title={documentTitle}>
        <Row>
            <Col lg={3}>
                <div className={"d-flex flex-row justify-content-center"}>
                    {me.Roles.includes(UserRole.ProductionTeam) ?
                        <ShowHideToggle
                            title="Unreleased"
                            state={showUnreleased}
                            setState={setShowUnreleased}/> : ""}

                    {me.Roles.includes(UserRole.ExecutiveDirector) ?
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
                    <PartsTopLinks downloadSession={downloadSession} project={project}/>
                    <PartsTable downloadSession={downloadSession} projectName={project.Name} parts={parts}/>
                </Col> :
                <Col>
                    <p>There are no projects currently accepting submissions, but we are working hard to bring you some!
                        <br/>Please check <LinkChannel channel={Channels.NextProjectHints}/> for updates.</p>
                </Col>}
        </Row>
    </RootContainer>;
};

const ButtonGroupBreakPoint = 800;

export const PartsTopLinks = (props: { downloadSession: Session, project: Project }) => {
    return <div className="d-flex justify-content-center">
        <ButtonGroup vertical={(window.visualViewport.width < ButtonGroupBreakPoint)}>
            <LinkButton to={links.RecordingInstructions}>
                <i className="far fa-image"/> Recording Instructions
            </LinkButton>
            <DownloadButton
                fileName={props.project.ReferenceTrack}
                downloadSession={props.downloadSession}>
                <i className="far fa-file-audio"/> Reference Track
            </DownloadButton>
            <LinkButton to={props.project.SubmissionLink}>
                <i className="fab fa-dropbox"/> Submit Recordings
            </LinkButton>
        </ButtonGroup>
    </div>;
};

const PartsTable = (props: { downloadSession: Session, projectName: string, parts: Part[] }) => {
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
                        <td><PartDownloads downloadSession={props.downloadSession} part={part}/></td>
                    </tr>)}
                </tbody>
            </Table>
        </div>
    </div>;
};

const PartDownloads = (props: { downloadSession: Session, part: Part }) => {
    const buttons = [] as Array<JSX.Element>;
    if (_.isEmpty(props.part.SheetMusicFile) == false)
        buttons.push(<DownloadButton
            key={props.part.SheetMusicFile}
            fileName={props.part.SheetMusicFile}
            downloadSession={props.downloadSession}
            size={"sm"}>
            <i className="far fa-file-pdf"/> sheet music
        </DownloadButton>);

    if (_.isEmpty(props.part.ClickTrackFile) == false)
        buttons.push(<DownloadButton
            key={props.part.ClickTrackFile}
            fileName={props.part.ClickTrackFile}
            downloadSession={props.downloadSession}
            size={"sm"}>
            <i className="far fa-file-audio"/> click track
        </DownloadButton>);

    if (_.isEmpty(props.part.ConductorVideo) == false)
        buttons.push(<LinkButton
            key={props.part.ConductorVideo}
            to={props.part.ConductorVideo}
            size={"sm"}>
            <i className="far fa-file-video"/> conductor video
        </LinkButton>);

    if (_.isEmpty(props.part.PronunciationGuide) == false)
        buttons.push(<DownloadButton
            key={props.part.PronunciationGuide}
            fileName={props.part.PronunciationGuide}
            downloadSession={props.downloadSession}
            size={"sm"}>
            <i className="fas fa-language"/> pronunciation guide
        </DownloadButton>);

    return <ButtonGroup
        className="justify-content-start"
        vertical={(window.visualViewport.width < ButtonGroupBreakPoint)}>
        {buttons}
    </ButtonGroup>;
};

const DownloadButton = (props: {
    downloadSession: Session,
    fileName: string,
    children: string | (string | JSX.Element)[]
    size?: "sm" | "lg"
}) => {
    const sessionKey = props.downloadSession ? props.downloadSession.Key : "";
    const params = new URLSearchParams({fileName: props.fileName, token: sessionKey});
    return <Button
        disabled={_.isEmpty(sessionKey)}
        href={"/download?" + params.toString()}
        variant="outline-light"
        size={props.size}>
        {props.children}
    </Button>;
};

const LinkButton = (props: {
    to: string
    children: string | (string | JSX.Element)[]
    size?: "sm" | "lg"
}) => {
    return <Button
        href={props.to}
        variant="outline-light"
        size={props.size}>
        {props.children}
    </Button>;
};
