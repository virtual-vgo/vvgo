import {MutableRefObject, useRef, useState} from "react";
import {Button, Card, Col, Dropdown, FormControl, InputGroup, Row, Table, Toast} from "react-bootstrap";
import {RootContainer} from "../components/shared/RootContainer";
import {GuildMember, useGuildMembers, useMixtapeProjects} from "../datasets";
import {deleteMixtapeProjects, MixtapeProject, saveMixtapeProjects} from "../datasets/MixtapeProject";
import React = require("react");

const GuildMemberToastLimit = 5;
const WintryMixChannelPrefix = "jackson-testing-";

export const NewProjectWorkflow = () => {
    const [projects, setProjects] = useMixtapeProjects();
    return <RootContainer>
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
    const onClickDelete = (project: MixtapeProject) => {
        console.log("deleting", project);
        props.setProjects(props.projects.filter(x => project.Id != x.Id));
        deleteMixtapeProjects([project]).catch(err => console.log(err));
    };
    return <Table className={"text-light"}>
        <tbody>
        {props.projects.map(x => <tr key={x.Id}>
            <td>{x.Mixtape}</td>
            <td>{x.Id}</td>
            <td>{x.Name}</td>
            <td>{x.Channel}</td>
            <td>{x.Owners.join(", ")}</td>
            <td><Button onClick={() => onClickDelete(x)}>Delete</Button></td>
        </tr>)}
        </tbody>
    </Table>;
};

const WorkflowApp = (props: {
    setProjects: (projects: MixtapeProject[]) => void;
    projects: MixtapeProject[];
}) => {
    const [curProject, setCurProject] = useState({Owners: []} as MixtapeProject);
    const saveProject = (project: MixtapeProject) => {
        props.setProjects([...props.projects.filter(p => p.Id != project.Id), project]);
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
                completed={curProject.Name && curProject.Name != ""}
            />
        </Col>
        <Col className={"col-md-6 mb-2"}>
            {curProject.Name && curProject.Name != "" ?
                <OwnersCard
                    title={"2. Choose project owners."}
                    saveProject={saveProject}
                    curProject={curProject}
                    limitResults={GuildMemberToastLimit}
                    completed={curProject.Owners && curProject.Owners.length > 0}
                /> : <div/>}
        </Col>
        <Col className={"col-md-6 mb-2"}>
            {curProject.Owners && curProject.Owners.length > 0 ?
                <ChannelCard
                    saveProject={saveProject}
                    curProject={curProject}
                    completed={curProject.Channel && curProject.Channel != ""}
                /> : <div/>}
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

    return <Card>
        <Card.Title className={"m-2 text-dark"}>
            <CheckMark completed={props.completed}>{props.title}</CheckMark>
        </Card.Title>
        <InputGroup>
            <InputGroup.Text>Name</InputGroup.Text>
            <FormControl
                ref={nameInputRef}
                onChange={() => setName(nameInputRef.current.value)}
                placeholder={"prOjEct NAmE"}
            />
            <SubmitNameButton
                nameInputRef={nameInputRef}
                current={props.curProject}
                saveProject={props.saveProject}
            />
        </InputGroup>
        <InputGroup>
            <InputGroup.Text>id</InputGroup.Text>
            <FormControl readOnly defaultValue={nameToId(name)}/>
        </InputGroup>;
    </Card>;
};

const SubmitNameButton = (props: {
    nameInputRef: MutableRefObject<HTMLInputElement>;
    current: MixtapeProject;
    saveProject: (project: MixtapeProject) => void;
}) => {
    const name = props.nameInputRef.current.value;
    const curId = props.current.Id;
    const id = curId && curId != "" ? curId : nameToId(name);
    return name && name != "" ?
        <Button
            variant={"outline-secondary"}
            onClick={() => props.saveProject({...props.current, Name: name, Id: id})}>
            Submit
        </Button> :
        <Button
            disabled
            variant={"outline-warning"}>
            required
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
    const [owners, setOwners] = useState(props.curProject.Owners.map(x => ({user: {id: x}} as GuildMember)));

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
            onChange={() => props.setSearchQuery(props.searchInputRef.current.value)}
        />
        <Button
            variant={"outline-secondary"}
            onClick={() => props.saveProject({...props.curProject, Owners: props.owners.map(x => x.user.id)})}>
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
    const guildMembers = useGuildMembers(props.searchQuery, props.limitResults);
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
                <Col className={"col-md-auto"}>3. Create a channel.</Col>
                <Col><CheckMark completed={props.completed}>3. Create a channel.</CheckMark>
                </Col>
            </Row>

        </Card.Title>
        <InputGroup>
            <InputGroup.Text>Channel</InputGroup.Text>
            <FormControl
                ref={channelInputRef}
                defaultValue={WintryMixChannelPrefix + props.curProject.Id}
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
        props.saveProject({...props.curProject, Channel: channel});
    };
    return <Button
        variant={"outline-secondary"}
        onClick={handleClick}>
        Submit
    </Button>;
};

const nameToId = (name: string): string =>
    name ? name.replace(/[ _]/, "-")
        .replace(/[^\w\d-]/g, "")
        .toLowerCase() : "";
