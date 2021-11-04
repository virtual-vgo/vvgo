import {useRef, useState} from "react";
import {Button, Card, Col, FormControl, InputGroup, Row} from "react-bootstrap";
import ReactMarkdown from "react-markdown";
import {getSession} from "../../auth";
import {links} from "../../data/links";
import {
    mixtapeProject,
    resolveHostNicks,
    saveMixtapeProjects,
    Session,
    useGuildMemberLookup,
    useMixtapeProjects,
    UserRole,
} from "../../datasets";
import {FancyProjectMenu, useMenuSelection} from "../shared/ProjectsMenu";
import {RootContainer} from "../shared/RootContainer";
import _ = require("lodash");
import React = require("react");

const pageTitle = "Wintry Mix | Members Dashboard";
const permaLink = (project: mixtapeProject) => `/mixtape/${project.Name}`;
const pathMatcher = /\/mixtape\/(.+)\/?/;

const searchProjects = (query: string, projects: mixtapeProject[]): mixtapeProject[] => {
    return _.defaultTo(projects, []).filter(project =>
        project.Name.toLowerCase().includes(query) ||
        project.channel.toLowerCase().includes(query) ||
        project.hosts.map(x => x.toLowerCase()).includes(query) ||
        project.tags.map(x => x.toLowerCase()).includes(query),
    );
};

export const MemberDashboard = () => {
    const [allProjects, setAllProjects] = useMixtapeProjects();
    const hosts = useGuildMemberLookup(allProjects.flatMap(r => r.hosts));
    const [selected, setSelected] = useMenuSelection(allProjects, pathMatcher, permaLink, _.shuffle(allProjects).pop());
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
                    choices={allProjects}
                    selected={selected}
                    setSelected={setSelected}
                    permaLink={permaLink}
                    searchChoices={searchProjects}
                    buttonContent={(proj) => <div>
                        {proj.title}<br/>
                        <small><em>{resolveHostNicks(hosts, proj).join(", ")}</em></small>
                    </div>}/>
            </Col>
            <Col lg={9}>
                {selected ?
                    <ProjectCard
                        me={me}
                        hostNicks={resolveHostNicks(hosts, selected)}
                        project={selected}
                        setProject={setSelected}
                        allProjects={allProjects}
                        setAllProjects={setAllProjects}/> :
                    <div/>}
            </Col>
        </Row>
    </RootContainer>;
};

const ProjectCard = (props: {
    me: Session
    hostNicks: string[]
    project: mixtapeProject
    setProject: (x: mixtapeProject) => void
    allProjects: mixtapeProject[]
    setAllProjects: (x: mixtapeProject[]) => void
}) => {
    const [showEdit, setShowEdit] = useState(false);
    const blurbRef = useRef({} as HTMLTextAreaElement);
    const tagsRef = useRef({} as HTMLInputElement);
    const canEdit = (props.me.DiscordID && props.project.hosts.includes(props.me.DiscordID)) ||
        (props.me.Roles && props.me.Roles.includes(UserRole.ExecutiveDirector));

    const onClickSubmit = () => {
        const update = {
            ...props.project,
            blurb: blurbRef.current.value,
            tags: tagsRef.current.value.split(",").map(t => t.trim()),
        };
        setShowEdit(false);
        saveMixtapeProjects([update])
            .then((resp) => {
                props.setProject(update);
                props.setAllProjects(_.uniqBy([...resp.MixtapeProjects, ...props.allProjects], x => x.Name));
            });

    };

    return <div>
        <h1>{props.project.title}</h1>
        <h4>
            Hosts: {props.hostNicks.join(", ")}<br/>
            Channel: <em>{props.project.channel}</em>
        </h4>
        {showEdit ?
            <div className="mb-3">
                <FormControl
                    ref={blurbRef}
                    as={"textarea"}
                    defaultValue={props.project.blurb}
                    placeholder={"Description"}/>
                <br/>
                <a href={links.Help.Markdown}>Markdown Cheatsheet</a>
            </div> :
            <ReactMarkdown>
                {props.project.blurb}
            </ReactMarkdown>}
        <Row>
            <Col>
                {showEdit ?
                    <InputGroup className="mb-3">
                        <InputGroup.Text>#</InputGroup.Text>
                        <FormControl
                            ref={tagsRef}
                            defaultValue={props.project.tags.join(", ")}
                            placeholder={"tags"}
                        />
                    </InputGroup> :
                    <Card.Text>
                        <i># {props.project.tags.join(", ")}</i>
                    </Card.Text>}
            </Col>
            {canEdit ?
                <Col className={"d-grid justify-content-end"}>
                    {showEdit ?
                        <Button
                            type={"button"}
                            variant={"outline-secondary"}
                            size={"sm"}
                            onClick={onClickSubmit}>
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
