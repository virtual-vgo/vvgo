import {isEmpty, uniq, uniqBy} from "lodash/fp";
import {MutableRefObject, useRef, useState} from "react";
import {Button, Card, Col, Dropdown, Form, FormControl, Row, Table, Toast} from "react-bootstrap";
import {
    GuildMember,
    MixtapeProject,
    useGuildMemberLookup,
    useGuildMemberSearch,
    useMixtapeProjects,
} from "../../datasets";
import {LoadingText} from "../shared/LoadingText";

const GuildMemberToastLimit = 5;
export const CurrentMixtape = "15b-wintry-mix";

export const NewProjectWorkflow = () => {
    const [projects, setProjects] = useMixtapeProjects();
    if (!projects) return <LoadingText/>;
    return <div>
        <div>
            <h1>Winter Mixtape</h1>
            <h2>New Project Workflow</h2>
            <WorkflowApp projects={projects} setProjects={setProjects}/>
            <h2>Existing Projects</h2>
            <ProjectTable projects={projects} setProjects={setProjects}/>
        </div>
    </div>;
};

const ProjectTable = (props: {
    projects: MixtapeProject[];
    setProjects: (projects: MixtapeProject[]) => void;
}) => {
    const hosts = useGuildMemberLookup(props.projects.flatMap(r => r.hosts ?? []));
    const onClickDelete = (project: MixtapeProject) => {
        console.log("deleting", project);
        props.setProjects(props.projects.filter(x => project.Name != x.Name));
        project.delete().catch(err => console.log(err));
    };

    const onClickEdit = (project: MixtapeProject) => {
        window.location.href = `/mixtape/NewProjectWorkflow?` + new URLSearchParams({name: project.Name});
    };

    return <Table className={"text-light"}>
        <tbody>
        {props.projects.map(x => <tr key={x.Name}>
            <td>{x.Name}</td>
            <td>{x.title}</td>
            <td>{x.channel}</td>
            <td>{x.resolveNicks(hosts).join(", ")}</td>
            <td><Button onClick={() => onClickDelete(x)}>Delete</Button></td>
            <td><Button onClick={() => onClickEdit(x)}>Edit</Button></td>
        </tr>)}
        </tbody>
    </Table>;
};

const WorkflowApp = (props: {
    setProjects: (projects: MixtapeProject[]) => void;
    projects: MixtapeProject[];
}) => {
    const params = new URLSearchParams(window.location.search);
    const initProject: MixtapeProject = props.projects
        .filter(x => x.Name == params.get("name")).pop() ?? new MixtapeProject("", CurrentMixtape);

    const [curProject, setCurProject] = useState(initProject);
    const saveProject = (project: MixtapeProject) => {
        props.setProjects(uniqBy(p => p.Name, [project, ...props.projects]));
        project.save()
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
                completed={!isEmpty(curProject.Name)}/>
        </Col>
        <Col className={"col-md-6 mb-2"}>
            {isEmpty(curProject.Name) ? <div/> :
                <SetHostsCard
                    cardTitle={"2. Choose project owners."}
                    saveProject={saveProject}
                    project={curProject}
                    toastLimit={GuildMemberToastLimit}
                    completed={!isEmpty(curProject.hosts)}/>}
        </Col>
        <Col className={"col-md-6 mb-2"}>
            {isEmpty(curProject.hosts) ? <div/> :
                <SetChannelCard
                    cardTitle={"3. Assign the project channel."}
                    saveProject={saveProject}
                    curProject={curProject}
                    completed={!isEmpty(curProject.channel)}/>}
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
    saveProject: (project: MixtapeProject) => void;
    curProject: MixtapeProject;
    completed: boolean;
}) => {
    const [title, setTitle] = useState(props.curProject.title ?? "");
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
    current: MixtapeProject
    saveProject: (project: MixtapeProject) => void
    className?: string
}) => {
    const name = props.current.Name != "" ? props.current.Name : nameToTitle(CurrentMixtape, props.title);
    const variant = isEmpty(props.title) ? "outline-warning" : "outline-primary";
    const onClick = () => {
        const proj = props.current;
        proj.Name = name;
        proj.title = props.title;
        props.saveProject(proj);
    };

    return <Button
        disabled={isEmpty(props.title)}
        variant={variant}
        className={props.className}
        onClick={onClick}>
        Submit
    </Button>;
};

const SetHostsCard = (props: {
    cardTitle: string;
    completed: boolean;
    project: MixtapeProject;
    saveProject: (project: MixtapeProject) => void;
    toastLimit: number;
}) => {
    const initHosts: GuildMember[] = props.project.hosts?.map(x => ({
        user: {id: x},
        nick: "",
        roles: [],
    })) ?? [];
    const [hosts, setHosts] = useState(initHosts);
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
                    {isEmpty(searchQuery) ? <div/> : <Toast>
                        {guildMembers
                            .filter(m => !isEmpty(m.nick))
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
    project: MixtapeProject
    saveProject: (proj: MixtapeProject) => void
    hosts: GuildMember[]
    setSearchQuery: (q: string) => void
}) => {
    const variant = isEmpty(props.hosts) ? "outline-warning" : "outline-primary";
    const onClick = () => {
        const proj = props.project;
        proj.hosts = uniq(props.hosts.map(x => x.user.id));
        props.setSearchQuery("");
        props.saveProject(proj);
    };
    return <Button
        disabled={isEmpty(props.hosts)}
        variant={variant}
        onClick={onClick}>
        Submit
    </Button>;
};

const SetChannelCard = (props: {
    cardTitle: string
    curProject: MixtapeProject;
    saveProject: (project: MixtapeProject) => void;
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
    curProject: MixtapeProject;
    saveProject: (project: MixtapeProject) => void;
}) => {
    const onClick = () => {
        const proj = props.curProject;
        proj.channel = props.channelInputRef.current.value;
        props.saveProject(proj);
    };
    return <Button
        variant={"outline-primary"}
        onClick={onClick}>
        Submit
    </Button>;
};

const nameToTitle = (mixtape: string, title: string | undefined): string => {
    const cleanTitle = (title ?? "")
        .replace(/[ _]/g, "-")
        .replace(/[^\w\d-]/g, "")
        .toLowerCase();

    return `${CurrentMixtape}-${cleanTitle}`;
};

export default NewProjectWorkflow;
