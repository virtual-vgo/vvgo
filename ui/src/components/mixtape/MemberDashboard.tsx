import {useRef, useState} from "react";
import {Button, Col, FormControl, Row} from "react-bootstrap";
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
    useProjects,
    UserRole,
} from "../../datasets";
import {FancyProjectMenu, useMenuSelection} from "../shared/ProjectsMenu";
import {RootContainer} from "../shared/RootContainer";
import {CurrentMixtape} from "./NewProjectWorkflow";
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
    const vvgoProjects = useProjects();
    const [mixtapeProjects, setMixtapeProjects] = useMixtapeProjects();
    const hosts = useGuildMemberLookup(mixtapeProjects.flatMap(r => r.hosts));
    const [selected, setSelected] = useMenuSelection(mixtapeProjects, pathMatcher, permaLink, _.shuffle(mixtapeProjects).pop());
    const me = getSession();

    const thisMixtape = _.defaultTo(vvgoProjects, []).filter(x => x.Name == CurrentMixtape).pop();

    return <RootContainer title={pageTitle}>
        <h1 className={"title"} style={{textAlign: "left"}}>
            Wintry Mix | Members Dashboard
        </h1>
        <h3>
            {_.isEmpty(thisMixtape) ? <div/> : <em>All submissions are due by {thisMixtape.SubmissionDeadline}.</em>}
        </h3>
        <Row className={"row-cols-1"}>
            <Col lg={3}>
                <FancyProjectMenu
                    choices={mixtapeProjects}
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
                        allProjects={mixtapeProjects}
                        setAllProjects={setMixtapeProjects}/> :
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
    const [showEdit, setShowEdit] = useState("");
    const blurbRef = useRef({} as HTMLTextAreaElement);
    const canEdit = (props.me.DiscordID && props.project.hosts.includes(props.me.DiscordID)) ||
        (props.me.Roles && props.me.Roles.includes(UserRole.ExecutiveDirector));

    const onClickSubmit = () => {
        const update = {
            ...props.project,
            blurb: blurbRef.current.value,
        };
        setShowEdit("");
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
        {showEdit == props.project.Name ?
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
        {canEdit ?
            showEdit == props.project.Name ?
                <Button
                    type={"button"}
                    variant={"outline-primary"}
                    size={"sm"}
                    onClick={onClickSubmit}>
                    Submit
                </Button> :
                <Button
                    type={"button"}
                    variant={"outline-primary"}
                    size={"sm"}
                    onClick={() => setShowEdit(props.project.Name)}>
                    Edit
                </Button> :
            <div/>}
    </div>;
};
