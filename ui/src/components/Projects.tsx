import React = require("react");
import Button from "react-bootstrap/Button";
import ButtonGroup from "react-bootstrap/ButtonGroup";
import Col from "react-bootstrap/Col";
import FormControl from "react-bootstrap/FormControl";
import Row from "react-bootstrap/Row";
import {latestProject, Project, useCreditsTable, useProjects} from "../datasets";
import {AlertUnreleasedProject} from "./shared/AlertUnreleasedProject";
import {LoadingText} from "./shared/LoadingText";
import {ProjectHeader} from "./shared/ProjectHeader";
import {RootContainer} from "./shared/RootContainer";
import {YoutubeIframe} from "./shared/YoutubeIframe";

export const Projects = () => {
    const allProjects = useProjects();

    const [project, setProject] = React.useState(null as Project);
    const [searchInput, setSearchInput] = React.useState("");
    const searchInputRef = React.useRef({} as HTMLInputElement);

    if (!allProjects) return <RootContainer>
        <LoadingText/>
    </RootContainer>;

    const wantProjects = allProjects.filter(r =>
        r.Name.toLowerCase().includes(searchInput) ||
        r.Title.toLowerCase().includes(searchInput) ||
        r.Sources.toLowerCase().includes(searchInput));

    const latest = latestProject(wantProjects);
    if (latest && !project) setProject(latest);

    return <RootContainer>
        <Row>
            <Col lg={3}>
                <div className="d-flex flex-row justify-content-center">
                    <FormControl
                        className="m-2"
                        ref={searchInputRef}
                        placeholder="search projects"
                        onChange={() => setSearchInput(searchInputRef.current.value.toLowerCase())}/>
                </div>
                <div className="d-flex flex-row justify-content-center">
                    <ButtonGroup vertical className="m-2">
                        {wantProjects.map(want =>
                            <Button
                                variant={project && project.Name == want.Name ? "light" : "outline-light"}
                                key={want.Name}
                                onClick={() => setProject(want)}>
                                {want.Title}
                                {want.PartsReleased == false ? <em><small><br/>Unreleased</small></em> : ""}
                                {want.VideoReleased == false ? <em><small><br/>In Production</small></em> : ""}
                                {want.VideoReleased == true ? <em><small><br/>Completed</small></em> : ""}
                            </Button>)}
                    </ButtonGroup>
                </div>
            </Col>
            <Col>
                {project ?
                    <div className="mx-4">
                        <AlertUnreleasedProject project={project}/>
                        <ProjectHeader project={project}/>
                        {project.YoutubeEmbed ?
                            <YoutubeIframe project={project}/> :
                            <div className="text-center">Video coming soon!</div>}
                        <ProjectCredits project={project}/>
                    </div> : <div/>}
            </Col>
        </Row>
    </RootContainer>;
};

const ProjectCredits = (props: { project: Project }) => {
    const creditsTable = useCreditsTable(props.project);

    if (creditsTable == null) return <div/>;
    return <div>
        {creditsTable.map(topic => <Row>
            <Row>
                <Col className="text-center">
                    <h2><strong>— {topic.Name} —</strong></h2>
                </Col>
            </Row>
            <Row>
                {topic.Rows.map(team => <Col lg={4}>
                    <h5>{team.Name}</h5>
                    <ul className="list-unstyled">
                        {team.Rows.map(credit =>
                            <li>{credit.Name} <small>{credit.BottomText}</small></li>)}
                    </ul>
                </Col>)}
            </Row>
        </Row>)}
    </div>;
};
