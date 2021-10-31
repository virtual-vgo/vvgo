import {useRef, useState} from "react";
import {Button, Card, Col, FormControl, InputGroup, Row} from "react-bootstrap";
import ButtonGroup from "react-bootstrap/ButtonGroup";
import ReactMarkdown from "react-markdown";
import {getSession} from "../../auth";
import {fetchApi, Session, useMixtapeProjects, UserRole} from "../../datasets";
import {MixtapeProject} from "../../datasets/MixtapeProject";
import {FancyProjectMenu, useMenuSelection} from "../shared/ProjectsMenu";
import {RootContainer} from "../shared/RootContainer";
import _ = require("lodash");
import React = require("react");

const pageTitle = "Wintry Mix | Members Dashboard";
const permaLink = (project: MixtapeProject) => `/mixtape/${project.Id}`;
const pathMatcher = /\/mixtape\/(.+)\/?/;

const searchProjects = (query: string, projects: MixtapeProject[]): MixtapeProject[] => {
    return _.defaultTo(projects, []).filter(project =>
        project.Name.toLowerCase().includes(query) ||
        project.Channel.toLowerCase().includes(query) ||
        project.Owners.map(x => x.toLowerCase()).includes(query) ||
        project.Tags.map(x => x.toLowerCase()).includes(query),
    );
};

export const MemberDashboard = () => {
    const [projects] = useMixtapeProjects();
    const shuffleProjects = _.shuffle(projects).map(p => {
        const tags = p.Tags ? p.Tags : [];
        const owners = p.Owners ? p.Owners : [];
        return {...p, Tags: tags, Owners: owners} as MixtapeProject;
    });
    const [selected, setSelected] = useMenuSelection(projects, pathMatcher, permaLink, _.shuffle(projects).pop());
    const me = getSession();

    return <RootContainer title={pageTitle}>
        <h1 className={"title"} style={{textAlign: "left"}}>
            Wintry Mix | Members Dashboard
        </h1>
        <h3>
            <em>All submissions are due by THIS DATE.</em>
        </h3>
        <Row className={"row-cols-1"}>
            <Col lg={3}>
                <FancyProjectMenu
                    choices={projects}
                    selected={selected}
                    setSelected={setSelected}
                    permaLink={null}
                    searchChoices={searchProjects}
                    buttonContent={(proj) =>
                        <div>
                            {proj.Name}
                            {proj.Owners ? <em><small>{proj.Owners.sort().join(", ")}</small></em> : ""}
                        </div>}/>
            </Col>
            <Col lg={9}>
                <Row className={"row-cols-1"}>
                    {shuffleProjects.map((p, i) =>
                        <Col key={i.toString()} className={"mt-3"}>
                            <ProjectCard me={me} project={p}/>
                        </Col>)}
                </Row>
            </Col>
        </Row>
    </RootContainer>;
};

const ProjectMenu = (props: {
    projects: MixtapeProject[],
    setSelected: (p: MixtapeProject) => void,
    selected: MixtapeProject,
}) => {
    const [searchInput, setSearchInput] = React.useState("");
    const searchInputRef = React.useRef({} as HTMLInputElement);
    const wantProjects = searchProjects(searchInput, props.projects);
    const onClickProject = (want: MixtapeProject) => {
        window.history.pushState(want, "", permaLink(want));
        props.setSelected(want);
    };

    return <div>
        <div className="d-flex flex-row justify-content-center">
            <FormControl
                className="m-2"
                ref={searchInputRef}
                placeholder="search projects"
                onChange={(event) => setSearchInput(event.target.value.toLowerCase())}/>
        </div>
        <div className="d-grid">
            <ButtonGroup vertical className="m-2">
                {wantProjects.map(want =>
                    <Button
                        variant={props.selected && props.selected.Name == want.Name ? "light" : "outline-light"}
                        key={want.Name}
                        onClick={() => onClickProject(want)}>
                        {want.Name}
                        <em><small><br/>{want.Owners.sort().join(", ")}</small></em>
                    </Button>)}
            </ButtonGroup>
        </div>
    </div>;
};

const ProjectCard = (props: { project: MixtapeProject, me: Session }) => {
    const {project, me} = props;

    const [showEdit, setShowEdit] = useState(false);
    const blurbRef = useRef({} as HTMLTextAreaElement);
    const tagsRef = useRef({} as HTMLInputElement);
    const canEdit = (me.DiscordID && project.Owners.includes(me.DiscordID)) ||
        (me.Roles && me.Roles.includes(UserRole.ExecutiveDirector));

    const onClickButton = () => {
        setShowEdit(false);
        project.Blurb = blurbRef.current.value;
        project.Tags = tagsRef.current.value.split(",").map(t => t.trim());
        fetchApi("/mixtape", {
            method: "POST",
            body: JSON.stringify([project]),
        }).then(resp => console.log(resp));
    };

    return <div>
        <h1>{project.Name}</h1>
        <h4>
            Hosts: {project.Owners.join(", ")}<br/>
            Channel: <em>{project.Channel}</em>
        </h4>
        {showEdit ?
            <InputGroup className="mb-3">
                <FormControl
                    ref={blurbRef}
                    as={"textarea"}
                    defaultValue={project.Blurb}
                    placeholder={"Description"}
                />
            </InputGroup> :
            <ReactMarkdown>
                {project.Blurb}
            </ReactMarkdown>}
        <Row>
            <Col>
                {showEdit ?
                    <InputGroup className="mb-3">
                        <InputGroup.Text>#</InputGroup.Text>
                        <FormControl
                            ref={tagsRef}
                            defaultValue={project.Tags.join(", ")}
                            placeholder={"tags"}
                        />
                    </InputGroup> :
                    <Card.Text>
                        <i># {project.Tags.join(", ")}</i>
                    </Card.Text>}
            </Col>
            {canEdit ?
                <Col className={"d-flex justify-content-end"}>
                    {showEdit ?
                        <Button
                            type={"button"}
                            variant={"outline-secondary"}
                            size={"sm"}
                            onClick={onClickButton}>
                            Submit
                        </Button> :
                        <Button
                            type={"button"}
                            variant={"outline-secondary"}
                            size={"sm"}
                            onClick={() => setShowEdit(true)}>
                            Edit
                        </Button>}
                </Col> : ""}
        </Row>
    </div>;
};
