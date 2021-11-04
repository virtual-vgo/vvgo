import {MutableRefObject, useRef, useState} from "react";
import {Button, Card, Col, Dropdown, Form, FormControl, Row, Table, Toast} from "react-bootstrap";
import {
    deleteMixtapeProjects,
    GuildMember,
    mixtapeProject,
    resolveHostNicks,
    saveMixtapeProjects,
    useGuildMemberLookup,
    useGuildMemberSearch,
    useMixtapeProjects,
} from "../../datasets";
import {LoadingText} from "../shared/LoadingText";
import {RootContainer} from "../shared/RootContainer";
import _ = require("lodash");
import React = require("react");

const GuildMemberToastLimit = 5;
export const CurrentMixtape = "15b-wintry-mix";

export const NewProjectWorkflow = () => {
    const [projects, setProjects] = useMixtapeProjects();
    return <RootContainer title="New Project Workflow">
        <h1>Winter Mixtape</h1>
        <h2>New Project Workflow</h2>
        <WorkflowApp projects={projects} setProjects={setProjects}/>
        <h2>Existing Projects</h2>
        <ProjectTable projects={projects} setProjects={setProjects}/>
    </RootContainer>;
};

const ProjectTable = (props: {
    projects: mixtapeProject[];
    setProjects: (projects: mixtapeProject[]) => void;
}) => {
    const hosts = useGuildMemberLookup(props.projects.flatMap(r => r.hosts));
    const onClickDelete = (project: mixtapeProject) => {
        console.log("deleting", project);
        props.setProjects(props.projects.filter(x => project.Name != x.Name));
        deleteMixtapeProjects([project]).catch(err => console.log(err));
    };

    if (!props.projects) return <LoadingText/>;

    return <Table className={"text-light"}>
        <tbody>
        {props.projects.map(x => <tr key={x.Name}>
            <td>{x.Name}</td>
            <td>{x.title}</td>
            <td>{x.channel}</td>
            <td>{resolveHostNicks(hosts, x).join(", ")}</td>
            <td><Button onClick={() => onClickDelete(x)}>Delete</Button></td>
        </tr>)}
        </tbody>
    </Table>;
};

const WorkflowApp = (props: {
    setProjects: (projects: mixtapeProject[]) => void;
    projects: mixtapeProject[];
}) => {
    const [curProject, setCurProject] = useState({mixtape: CurrentMixtape, hosts: []} as mixtapeProject);
    const saveProject = (project: mixtapeProject) => {
        props.setProjects(_.uniqBy([project, ...props.projects], x => x.Name));
        saveMixtapeProjects([project])
            .then(() => setCurProject(project))
            .catch(err => console.log(err));
        return;
    };

    return <Row className={"row-cols-1"}>
        <Col className={"col-md-6 mb-2"}>
            <SetTitleCard
                cardTitle={"1. Create the project."}
                saveProject={saveProject}
                curProject={curProject}
                completed={!_.isEmpty(curProject.Name)}/>
        </Col>
        <Col className={"col-md-6 mb-2"}>
            {_.isEmpty(curProject.Name) ? <div/> :
                <SetHostsCard
                    cardTitle={"2. Choose project owners."}
                    saveProject={saveProject}
                    project={curProject}
                    toastLimit={GuildMemberToastLimit}
                    completed={curProject.hosts && curProject.hosts.length > 0}/>}
        </Col>
        <Col className={"col-md-6 mb-2"}>
            {_.isEmpty(curProject.hosts) ? <div/> :
                <SetChannelCard
                    cardTitle={"3. Create the project channel."}
                    saveProject={saveProject}
                    curProject={curProject}
                    completed={curProject.channel && curProject.channel != ""}/>}
        </Col>
    </Row>;
};

const CheckMark = (props: {
    children: JSX.Element[] | JSX.Element | string;
    completed: boolean
}) => <Row>
    <Col>{props.children}</Col>
    <Col md={"auto"} className={"text-end"}>{props.completed ? "✔️" : ""}</Col>
</Row>;

const SetTitleCard = (props: {
    cardTitle: string
    saveProject: (project: mixtapeProject) => void;
    curProject: mixtapeProject;
    completed: boolean;
}) => {
    const [title, setTitle] = useState(props.curProject.title);
    return <Card className={"bg-transparent"}>
        <Card.Title className={"m-2"}>
            <CheckMark completed={props.completed}>{props.cardTitle}</CheckMark>
        </Card.Title>
        <Card.Body>
            <Form>
                <Form.Group>
                    <Form.Label>Title</Form.Label>
                    <FormControl
                        onChange={(event) => setTitle(event.target.value)}
                        placeholder={"prOjEct TItlE"}/>
                </Form.Group>
                <Form.Group>
                    <Form.Label>Name</Form.Label>
                    <FormControl readOnly value={nameToTitle(CurrentMixtape, title)}/>
                </Form.Group>
                <SubmitTitleButton
                    title={title}
                    current={props.curProject}
                    saveProject={props.saveProject}/>
            </Form>
        </Card.Body>
    </Card>;
};

const SubmitTitleButton = (props: {
    title: string
    current: mixtapeProject
    saveProject: (project: mixtapeProject) => void
    className?: string
}) => {
    const name = _.defaultTo(props.current.Name, nameToTitle(CurrentMixtape, props.title));
    const variant = _.isEmpty(props.title) ? "outline-warning" : "outline-primary";
    return <Button
        disabled={_.isEmpty(props.title)}
        variant={variant}
        className={props.className}
        onClick={() => props.saveProject({
            ...props.current,
            Name: name,
            title: props.title,
        })}>
        Submit
    </Button>;
};

const SetHostsCard = (props: {
    cardTitle: string;
    completed: boolean;
    project: mixtapeProject;
    saveProject: (project: mixtapeProject) => void;
    toastLimit: number;
}) => {
    const [hosts, setHosts] = useState(props.project.hosts.map(x => ({user: {id: x}} as GuildMember)));
    const [searchQuery, setSearchQuery] = useState("");
    const guildMembers = useGuildMemberSearch(searchQuery, props.toastLimit);

    return <Card
        className={"bg-transparent"}
        onKeyUp={(e) => {
            if (e.key == "Escape") setSearchQuery("");
        }}>
        <Card.Title>
            <CheckMark completed={props.completed}>{props.cardTitle}</CheckMark>
        </Card.Title>
        <Card.Body>
            <Form>
                Hosts<br/>
                <ul className="list-inline">
                    {hosts.map(host => <li key={host.nick} className="list-inline-item">
                        {host.nick + " "}
                        <span
                            className={"text-primary"}
                            style={{cursor: "pointer"}}
                            onClick={() => setHosts(hosts.filter(x => x.user.id != host.user.id))}>
                            (remove)
                        </span>
                    </li>)}
                </ul>
                <Form.Group>
                    <Form.Control
                        placeholder="search nicks"
                        defaultValue={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}/>
                    {_.isEmpty(searchQuery) ? <div/> : <Toast>
                        {guildMembers
                            .filter(m => !_.isEmpty(m.nick))
                            .filter(m => !hosts.includes(m))
                            .map(m => <Dropdown.Item
                                key={m.nick}
                                onClick={() => setHosts([...hosts, m])}>
                                {m.nick}
                            </Dropdown.Item>)}
                    </Toast>}
                </Form.Group>
                <SubmitHostsButton
                    project={props.project}
                    saveProject={props.saveProject}
                    hosts={hosts}
                    setSearchQuery={setSearchQuery}/>
            </Form>
        </Card.Body>
    </Card>;
};

const SubmitHostsButton = (props: {
    project: mixtapeProject
    saveProject: (proj: mixtapeProject) => void
    hosts: GuildMember[]
    setSearchQuery: (q: string) => void
}) => {
    const variant = _.isEmpty(props.hosts) ? "outline-warning" : "outline-primary";
    return <Button
        disabled={_.isEmpty(props.hosts)}
        variant={variant}
        onClick={() => {
            props.setSearchQuery("");
            props.saveProject({
                ...props.project,
                hosts: _.uniq(props.hosts.map(x => x.user.id)),
            });
        }}>
        Submit
    </Button>;
};

const SetChannelCard = (props: {
    cardTitle: string
    curProject: mixtapeProject;
    saveProject: (project: mixtapeProject) => void;
    completed: boolean;
}) => {
    const channelInputRef = useRef({} as HTMLInputElement);
    return <Card className={"bg-transparent"}>
        <Card.Title>
            <CheckMark completed={props.completed}>{props.cardTitle}</CheckMark>
        </Card.Title>
        <Card.Body>
            <Form>
                <Form.Group>
                    <Form.Label>Channel</Form.Label>
                    <Form.Control
                        ref={channelInputRef}
                        defaultValue={props.curProject.Name}/>
                </Form.Group>
                <SubmitChannelButton
                    channelInputRef={channelInputRef}
                    curProject={props.curProject}
                    saveProject={props.saveProject}/>
            </Form>
        </Card.Body>
    </Card>;
};

const SubmitChannelButton = (props: {
    channelInputRef: MutableRefObject<HTMLInputElement>
    curProject: mixtapeProject;
    saveProject: (project: mixtapeProject) => void;
}) => {
    const handleClick = () => {
        const channel = props.channelInputRef.current.value;
        props.saveProject({...props.curProject, channel: channel});
    };
    return <Button
        variant={"outline-primary"}
        onClick={handleClick}>
        Submit
    </Button>;
};

const nameToTitle = (mixtape: string, title: string): string => {
    const cleanTitle = _.defaultTo(title, "")
        .replace(/[ _]/g, "-")
        .replace(/[^\w\d-]/g, "")
        .toLowerCase();

    return `${CurrentMixtape}-${cleanTitle}`;
};
