import {MutableRefObject, useRef, useState} from "react";
import {Button, Card, Col, Dropdown, FormControl, InputGroup, Row, Table, Toast} from "react-bootstrap";
import {
    deleteMixtapeProjects,
    GuildMember,
    MixtapeProject,
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
const WintryMixChannelPrefix = "jackson-testing-";

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
    projects: MixtapeProject[];
    setProjects: (projects: MixtapeProject[]) => void;
}) => {
    const hosts = useGuildMemberLookup(props.projects.flatMap(r => r.hosts));
    const onClickDelete = (project: MixtapeProject) => {
        console.log("deleting", project);
        props.setProjects(props.projects.filter(x => project.id != x.id));
        deleteMixtapeProjects([project]).catch(err => console.log(err));
    };

    if (!props.projects) return <LoadingText/>;

    return <Table className={"text-light"}>
        <tbody>
        {props.projects.map(x => <tr key={x.id}>
            <td>{x.mixtape}</td>
            <td>{x.id}</td>
            <td>{x.Name}</td>
            <td>{x.channel}</td>
            <td>{resolveHostNicks(hosts, x).join(", ")}</td>
            <td><Button onClick={() => onClickDelete(x)}>Delete</Button></td>
        </tr>)}
        </tbody>
    </Table>;
};

const WorkflowApp = (props: {
    setProjects: (projects: MixtapeProject[]) => void;
    projects: MixtapeProject[];
}) => {
    const [curProject, setCurProject] = useState({hosts: []} as MixtapeProject);
    const saveProject = (project: MixtapeProject) => {
        props.setProjects([...props.projects.filter(p => p.id != project.id), project]);
        saveMixtapeProjects([project])
            .then(() => setCurProject(project))
            .catch(err => console.log(err));
        return;
    };

    return <Row className={"row-cols-1"}>
        <Col className={"col-md-6 mb-2"}>
            <NameCard
                title={"1. Set the project name."}
                saveProject={saveProject}
                curProject={curProject}
                completed={!_.isEmpty(curProject.Name)}
            />
        </Col>
        <Col className={"col-md-6 mb-2"}>
            {_.isEmpty(curProject.Name) ? <div/> :
                <OwnersCard
                    title={"2. Choose project owners."}
                    saveProject={saveProject}
                    curProject={curProject}
                    limitResults={GuildMemberToastLimit}
                    completed={curProject.hosts && curProject.hosts.length > 0}
                />}
        </Col>
        <Col className={"col-md-6 mb-2"}>
            {_.isEmpty(curProject.hosts) ? <div/> :
                <ChannelCard
                    saveProject={saveProject}
                    curProject={curProject}
                    completed={curProject.channel && curProject.channel != ""}
                />}
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

const NameCard = (props: {
    title: string
    saveProject: (project: MixtapeProject) => void;
    curProject: MixtapeProject;
    completed: boolean;
}) => {
    const nameInputRef = useRef({} as HTMLInputElement);
    const [name, setName] = useState(props.curProject.Name);

    return <Card className={"bg-transparent"}>
        <Card.Title className={"m-2"}>
            <CheckMark completed={props.completed}>{props.title}</CheckMark>
        </Card.Title>
        <InputGroup>
            <InputGroup.Text>Name</InputGroup.Text>
            <FormControl
                ref={nameInputRef}
                onChange={(event) => setName(event.target.value)}
                placeholder={"prOjEct NAmE"}
            />
            <SubmitNameButton
                name={name}
                current={props.curProject}
                saveProject={props.saveProject}
            />
        </InputGroup>
        <InputGroup>
            <InputGroup.Text>id</InputGroup.Text>
            <FormControl readOnly defaultValue={nameToId(name)}/>
        </InputGroup>
    </Card>;
};

const SubmitNameButton = (props: {
    name: string;
    current: MixtapeProject;
    saveProject: (project: MixtapeProject) => void;
}) => {
    const curId = props.current.id;
    const id = curId && curId != "" ? curId : nameToId(props.name);
    return _.isEmpty(props.name) ?
        <Button
            disabled
            variant={"outline-warning"}>
            required
        </Button> :
        <Button
            variant={"outline-secondary"}
            onClick={() => props.saveProject({...props.current, Name: props.name, id: id})}>
            Submit
        </Button>;
};

const OwnersCard = (props: {
    title: string;
    limitResults: number;
    curProject: MixtapeProject;
    saveProject: (project: MixtapeProject) => void;
    completed: boolean;
}) => {
    const searchInputRef = useRef({} as HTMLInputElement);
    const [searchQuery, setSearchQuery] = useState("");
    const [owners, setOwners] = useState(props.curProject.hosts.map(x => ({user: {id: x}} as GuildMember)));

    return <Card>
        <Card.Title className={"m-2 text-dark"}>
            <CheckMark completed={props.completed}>{props.title}</CheckMark>
        </Card.Title>
        <EditMembers
            owners={owners}
            setOwners={setOwners}
            saveProject={props.saveProject}
        />
        <SearchMembers
            searchInputRef={searchInputRef}
            searchQuery={searchQuery}
            setSearchQuery={setSearchQuery}
            curProject={props.curProject}
            owners={owners}
            saveProject={props.saveProject}
        />
        <MembersToast
            searchInputRef={searchInputRef}
            searchQuery={searchQuery}
            owners={owners}
            setOwners={setOwners}
            limitResults={props.limitResults}
        />
    </Card>;
};

const EditMembers = (props: {
    owners: GuildMember[];
    setOwners: (owners: GuildMember[]) => void;
    saveProject: (project: MixtapeProject) => void;
}) => {
    return <InputGroup>
        <InputGroup.Text>Owners</InputGroup.Text>
        {props.owners.map(owner =>
            <Button
                key={owner.nick}
                variant={"outline-primary"}
                onClick={() => props.setOwners(props.owners.filter(x => x.user != owner.user))}>
                {owner.nick}
            </Button>,
        )}
    </InputGroup>;
};

const SearchMembers = (props: {
    searchInputRef: MutableRefObject<HTMLInputElement>;
    searchQuery: string;
    setSearchQuery: (x: string) => void;
    saveProject: (proj: MixtapeProject) => void;
    owners: GuildMember[];
    curProject: MixtapeProject;
}) => {
    return <InputGroup>
        <InputGroup.Text>Search</InputGroup.Text>
        <FormControl
            ref={props.searchInputRef}
            placeholder="search nicks"
            defaultValue={props.searchQuery}
            onChange={() => props.setSearchQuery(props.searchInputRef.current.value)}/>
        <Button
            variant={"outline-secondary"}
            onClick={() => props.saveProject({...props.curProject, hosts: props.owners.map(x => x.user.id)})}>
            Submit
        </Button>
    </InputGroup>;
};

const MembersToast = (props: {
    searchInputRef: MutableRefObject<HTMLInputElement>;
    searchQuery: string;
    owners: GuildMember[];
    setOwners: (x: GuildMember[]) => void;
    limitResults: number;
}) => {
    const guildMembers = useGuildMemberSearch(props.searchQuery, props.limitResults);
    return <Toast>
        {props.searchInputRef.current.value === "" ? "" :
            guildMembers
                .filter(m => m.nick)
                .filter(m => m.nick !== "")
                .filter(m => !props.owners.includes(m))
                .map(m => <Dropdown.Item
                    key={m.nick}
                    onClick={() => props.setOwners([...props.owners, m])}>
                    {m.nick}
                </Dropdown.Item>)}
    </Toast>;
};

const ChannelCard = (props: {
    curProject: MixtapeProject;
    saveProject: (project: MixtapeProject) => void;
    completed: boolean;
}) => {
    const channelInputRef = useRef({} as HTMLInputElement);
    return <Card>
        <Card.Title className={"m-2 text-dark"}>
            <Row>
                <Col>
                    <CheckMark completed={props.completed}>
                        3. Create a channel.
                    </CheckMark>
                </Col>
            </Row>
        </Card.Title>
        <InputGroup>
            <InputGroup.Text>Channel</InputGroup.Text>
            <FormControl
                ref={channelInputRef}
                defaultValue={WintryMixChannelPrefix + props.curProject.id}
            />
            <SubmitChannelButton
                channelInputRef={channelInputRef}
                curProject={props.curProject}
                saveProject={props.saveProject}
            />
        </InputGroup>
    </Card>;
};

const SubmitChannelButton = (props: {
    channelInputRef: MutableRefObject<HTMLInputElement>
    curProject: MixtapeProject;
    saveProject: (project: MixtapeProject) => void;
}) => {
    const handleClick = () => {
        const channel = props.channelInputRef.current.value;
        props.saveProject({...props.curProject, channel: channel});
    };
    return <Button
        variant={"outline-secondary"}
        onClick={handleClick}>
        Submit
    </Button>;
};

const nameToId = (name: string): string => _.isEmpty(name) ?
    "" :
    name.replace(/[ _]/g, "-")
        .replace(/[^\w\d-]/g, "")
        .toLowerCase();
