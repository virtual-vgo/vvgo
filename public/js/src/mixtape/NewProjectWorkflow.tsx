import {MutableRefObject, useRef, useState} from "react";
import {Button, Card, Col, FormControl, InputGroup, Row, Toast} from "react-bootstrap";
import {Container} from "../components";
import {GuildMember, useGuildMembers} from "../datasets";
import {MixtapeProject, saveMixtapeProject} from "../datasets/MixtapeProject";
import React = require("react");

export const NewProjectWorkflow = () => {
    const limit = 5;
    const [name, setName] = useState("");
    const [query, setQuery] = useState("");
    const [owners, setOwners] = useState([] as GuildMember[]);
    const guildMembers = useGuildMembers(query, limit);

    return <Container>
        <h1>Winter Mixtape</h1>
        <h2>New Project Workflow</h2>
        <Row>
            <Col className={"col-md-6"}>
                <CreateProjectForm
                    name={name}
                    setName={setName}
                    owners={owners}
                    setOwners={setOwners}
                    setQuery={setQuery}
                    guildMembers={guildMembers}/>
            </Col>
        </Row>
    </Container>;
};

const CreateProjectForm = (props: {
    name: string;
    guildMembers: GuildMember[];
    owners: GuildMember[];
    setName: (x: string) => void;
    setQuery: (x: string) => void;
    setOwners: (x: GuildMember[]) => void;
}) => {
    const {guildMembers, owners, setQuery, setOwners, name, setName} = props;
    const nameInputRef = useRef({} as HTMLInputElement);
    const searchInputRef = useRef({} as HTMLInputElement);
    const [newProject, setNewProject] = useState({} as MixtapeProject);

    const saveNewProject = () => {
        const proj: MixtapeProject = {
            Name: name,
            Channel: "",
            Owners: owners.map(x => x.user),
            Blurb: "",
            Tags: [],
        };
        saveMixtapeProject(proj)
            .then(() => setNewProject(proj))
            .catch(err => console.log(err));
    };

    return <Card>
        <InputName
            name={name}
            setName={setName}
            nameInputRef={nameInputRef}
            newProject={newProject}
            saveNewProject={saveNewProject}
        />
        <InputId name={name}/>
        <ShowOwners owners={owners} setOwners={setOwners}/>
        <SearchMembers searchInputRef={searchInputRef} setQuery={setQuery}/>
        <Toast>
            {searchInputRef.current.value === "" ? "" :
                guildMembers
                    .filter(m => m.nick)
                    .filter(m => m.nick !== "")
                    .filter(m => !owners.includes(m))
                    .map(m => <Toast.Body
                        key={m.nick}
                        children={m.nick}
                        onClick={() => setOwners([...owners, m])}
                    />)}
        </Toast>
    </Card>;
};

const InputName = (props: {
    name: string;
    nameInputRef: MutableRefObject<HTMLInputElement>;
    setName: (name: string) => void;
    newProject: MixtapeProject;
    saveNewProject: () => void;
}) => {

    const SubmitButton = () => {
        const value = props.nameInputRef.current.value;
        switch (true) {
            case value == "" || !value:
                return <Button
                    variant={"outline-warning"}
                    children={"required"}
                />;
            case value == props.newProject.Name:
                return <Button
                    disabled
                    variant={"outline-success"}
                    children={"✔️"}
                />;
            default:
                return <Button
                    variant={"outline-secondary"}
                    children={"Submit"}
                    onClick={() => props.saveNewProject()}
                />;
        }
    };

    return <InputGroup>
        <InputGroup.Text>Name</InputGroup.Text>
        <FormControl
            ref={props.nameInputRef}
            onChange={() => props.setName(props.nameInputRef.current.value)}
            defaultValue={props.newProject.Name}
            placeholder={"prOjEct NAmE"}
        />
        <SubmitButton/>
    </InputGroup>;
};

const InputId = (props: { name: string }) => {
    const {name} = props;
    return <InputGroup>
        <InputGroup.Text>id</InputGroup.Text>
        <FormControl
            readOnly
            defaultValue={nameToId(name)}/>
    </InputGroup>;
};

const ShowOwners = (props: {
    owners: GuildMember[];
    setOwners: (owners: GuildMember[]) => void;
}) => {
    const {owners, setOwners} = props;
    return <InputGroup>
        <InputGroup.Text>Owners</InputGroup.Text>
        {owners.map(owner =>
            <Button
                variant={"primary"}
                children={owner.nick}
                onClick={() => setOwners(owners.filter(x => x.user != owner.user))}
            />,
        )}
        <Button
            variant={"outline-secondary"}
            children={"Submit"}/>
    </InputGroup>;
};

const SearchMembers = (props: {
    searchInputRef: MutableRefObject<HTMLInputElement>;
    setQuery: (query: string) => void;
}) => {
    const {searchInputRef, setQuery} = props;
    return <InputGroup>
        <InputGroup.Text>Search</InputGroup.Text>
        <FormControl
            ref={searchInputRef}
            placeholder="search nicks"
            onChange={() => setQuery(searchInputRef.current.value)}
        />
    </InputGroup>;
};

const nameToId = (name: string): string =>
    name.replace(/[ _]/, "-").toLowerCase();
