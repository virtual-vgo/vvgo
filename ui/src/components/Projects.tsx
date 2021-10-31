import _ = require("lodash");
import React = require("react");
import Masonry from "@mui/lab/Masonry";
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

export const searchProjects = (query: string, projects: Project[]): Project[] => {
    return _.defaultTo(projects, []).filter(r =>
        r.Name.toLowerCase().includes(query) ||
        r.Title.toLowerCase().includes(query) ||
        r.Sources.toLowerCase().includes(query));
};

export const Projects = () => {
    const documentTitle = "Projects";
    const allProjects = useProjects();
    const [selected, setSelected] = React.useState(null as Project);

    if (!allProjects) return <RootContainer title={documentTitle}>
        <LoadingText/>
    </RootContainer>;

    const allowedProjects = allProjects.filter(r => r.Hidden == false);
    initializeSelected(selected, setSelected, allowedProjects);
    return <RootContainer title={documentTitle}>
        <Row>
            <Col lg={3}>
                <ProjectMenu
                    projects={allowedProjects}
                    setSelected={setSelected}
                    selected={selected}/>
            </Col>
            <Col>
                {selected ?
                    <div className="mx-4">
                        <AlertUnreleasedProject project={selected}/>
                        <ProjectHeader project={selected}/>
                        {selected.PartsArchived ?
                            selected.YoutubeEmbed ?
                                <YoutubeIframe project={selected}/> :
                                <div className="text-center text-info">
                                    <em>Video coming soon!</em>
                                </div> :
                            <div/>}
                        <ProjectCredits project={selected}/>
                    </div> :
                    <div/>}
            </Col>
        </Row>
    </RootContainer>;
};

const initializeSelected = (selected: Project, setSelected: (p: Project) => void, projects: Project[]) => {
    window.onpopstate = (e) => {
        const params = new URLSearchParams(e.state);
        const want = projects.filter(r => r.Name == params.get("name")).pop();
        if (want) setSelected(want);
    };

    if (!selected) { // initialize from url query or from latest project
        const params = new URLSearchParams(document.location.search);
        if (!_.isEmpty(params.get("name"))) {
            const want = projects.filter(r => r.Name == params.get("name")).pop();
            if (want) setSelected(want);
            window.history.pushState(params, "", "/projects?" + params.toString());
        } else {
            const latest = latestProject(projects);
            if (latest) setSelected(latest);
        }
    }
};

const ProjectMenu = (props: {
    projects: Project[],
    setSelected: (p: Project) => void,
    selected: Project,
}) => {
    const [searchInput, setSearchInput] = React.useState("");
    const searchInputRef = React.useRef({} as HTMLInputElement);
    const wantProjects = searchProjects(searchInput, props.projects);
    const onClickProject = (want: Project) => {
        const params = new URLSearchParams({name: want.Name});
        window.history.pushState(params, "", "/projects?" + params.toString());
        props.setSelected(want);
    };

    return <div>
        <div className="d-flex flex-row justify-content-center">
            <FormControl
                className="m-2"
                ref={searchInputRef}
                placeholder="search projects"
                onChange={() => setSearchInput(searchInputRef.current.value.toLowerCase())}/>
        </div>
        <div className="d-grid">
            <ButtonGroup vertical className="m-2">
                {wantProjects.map(want =>
                    <Button
                        variant={props.selected && props.selected.Name == want.Name ? "light" : "outline-light"}
                        key={want.Name}
                        onClick={() => onClickProject(want)}>
                        {want.Title}
                        {want.PartsReleased == false ? <em><small><br/>Unreleased</small></em> : ""}
                        {want.VideoReleased == false ? <em><small><br/>In Production</small></em> : ""}
                        {want.VideoReleased == true ? <em><small><br/>Completed</small></em> : ""}
                    </Button>)}
            </ButtonGroup>
        </div>
    </div>;
};

const ProjectCredits = (props: { project: Project }) => {
    const creditsTable = useCreditsTable(props.project);

    if (_.isEmpty(creditsTable)) return <div/>;
    return <div>
        {creditsTable.map(topic => <Row key={topic.Name}>
            <Row>
                <Col className="text-center">
                    <h2><strong>— {topic.Name} —</strong></h2>
                </Col>
            </Row>
            <Row>
                <Masonry
                    columns={3}
                    spacing={1}
                    defaultHeight={450}
                    defaultColumns={3}
                    defaultSpacing={1}>
                    {_.isEmpty(topic.Rows) ? <div/> : topic.Rows.map(team =>
                        <Col key={team.Name} lg={4}>
                            <h5>{team.Name}</h5>
                            <ul className="list-unstyled">
                                {team.Rows.map(credit =>
                                    <li key={credit.Name}>{credit.Name} <small>{credit.BottomText}</small></li>)}
                            </ul>
                        </Col>)}
                </Masonry>
            </Row>
        </Row>)}
    </div>;
};

