import {useRef, useState} from "react";
import {Button, Card, Col, FormControl, InputGroup, Row} from "react-bootstrap";
import ReactMarkdown from "react-markdown";
import {getSession} from "../auth";
import {RootContainer} from "../components";
import {fetchApi, Roles, Session, useMixtapeProjects} from "../datasets";
import {MixtapeProject} from "../datasets/MixtapeProject";
import _ = require("lodash");
import React = require("react");

export const MemberDashboard = () => {
    const [projects] = useMixtapeProjects();
    const shuffleProjects = _.shuffle(projects).map(p => {
        const tags = p.Tags ? p.Tags : [];
        const owners = p.Owners ? p.Owners : [];
        return {...p, Tags: tags, Owners: owners} as MixtapeProject;
    });
    const me = getSession();
    return <RootContainer>
        <Row className={"row-cols-1"}>
            <Col>
                <h1 className={"title"} style={{textAlign: "left"}}>Wintry Mix | Members Dashboard</h1>
            </Col>
            <Col className={"mt-3"}>
                <Row md={2} sm={1}>
                    {shuffleProjects.map((p, i) =>
                        <Col key={i.toString()} className={"mt-3"}>
                            <ProjectCard me={me} project={p}/>
                        </Col>)}
                </Row>
            </Col>
        </Row>
    </RootContainer>;
};

const ProjectCard = (props: { project: MixtapeProject, me: Session }) => {
    const {project, me} = props;

    const [showEdit, setShowEdit] = useState(false);
    const blurbRef = useRef({} as HTMLTextAreaElement);
    const tagsRef = useRef({} as HTMLInputElement);
    const canEdit = (me.DiscordID && project.Owners.includes(me.DiscordID)) ||
        (me.Roles && me.Roles.includes(Roles.ExecutiveDirector));

    const buttonOnClick = () => {
        setShowEdit(false);
        project.Blurb = blurbRef.current.value;
        project.Tags = tagsRef.current.value.split(",").map(t => t.trim());
        fetchApi("/mixtape", {
            method: "POST",
            body: JSON.stringify([project]),
        }).then(resp => console.log(resp));
    };

    return <Card>
        <Card.Body className={"text-dark"}>
            <Card.Title>{project.Name}</Card.Title>
            <Card.Subtitle className="mb-2 text-muted">
                Project Owners: {project.Owners.join(", ")}<br/>
                Channel: {project.Channel}
            </Card.Subtitle>
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
                                onClick={buttonOnClick}>
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
        </Card.Body>
    </Card>;
};
