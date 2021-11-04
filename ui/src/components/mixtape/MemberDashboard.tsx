import {useRef, useState} from "react";
import {Button, Card, Col, FormControl, InputGroup, Row} from "react-bootstrap";
import ReactMarkdown from "react-markdown";
import {getSession} from "../../auth";
import {
    fetchApi,
    mixtapeProject,
    resolveHostNicks,
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
    const [projects] = useMixtapeProjects();
    const hosts = useGuildMemberLookup(projects.flatMap(r => r.hosts));
    const shuffleProjects = _.shuffle(projects).map(p => {
        return {...p, tags: _.defaultTo(p.tags, []), hosts: _.defaultTo(p.hosts, [])} as mixtapeProject;
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
                    permaLink={permaLink}
                    searchChoices={searchProjects}
                    buttonContent={(proj) => <div>
                        {proj.title}<br/>
                        <small><em>{resolveHostNicks(hosts, proj).join(", ")}</em></small>
                    </div>}/>
            </Col>
            <Col lg={9}>
                <ProjectCard me={me} hostNicks={resolveHostNicks(hosts, selected)} project={selected}/>
            </Col>
        </Row>
    </RootContainer>;
};

const ProjectCard = (props: { project: mixtapeProject, hostNicks: string[], me: Session }) => {
    const [showEdit, setShowEdit] = useState(false);
    const blurbRef = useRef({} as HTMLTextAreaElement);
    const tagsRef = useRef({} as HTMLInputElement);
    const canEdit = (props.me.DiscordID && props.project.hosts.includes(props.me.DiscordID)) ||
        (props.me.Roles && props.me.Roles.includes(UserRole.ExecutiveDirector));

    const onClickButton = () => {
        setShowEdit(false);
        fetchApi("/mixtape", {
            method: "POST",
            body: JSON.stringify([{
                ...props.project,
                blurb: blurbRef.current.value,
                tags: tagsRef.current.value.split(",").map(t => t.trim()),
            }]),
        }).then(resp => console.log(resp));
    };

    return <div>
        <h1>{props.project.title}</h1>
        <h4>
            Hosts: {props.hostNicks.join(", ")}<br/>
            Channel: <em>{props.project.channel}</em>
        </h4>
        {showEdit ?
            <InputGroup className="mb-3">
                <FormControl
                    ref={blurbRef}
                    as={"textarea"}
                    defaultValue={props.project.blurb}
                    placeholder={"Description"}
                />
            </InputGroup> :
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
