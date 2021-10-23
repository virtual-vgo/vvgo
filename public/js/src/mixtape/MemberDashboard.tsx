import {useRef, useState} from "react";
import {Button, Card, Col, FormControl, InputGroup, Row} from "react-bootstrap";
import ReactMarkdown from "react-markdown";
import {Container} from "../components";
import {fetchApi, Session, useMixtapeProjects, useMySession} from "../datasets";
import {MixtapeProject} from "../datasets/MixtapeProject";
import _ = require("lodash");
import React = require("react");

export const MemberDashboard = () => {
    const projects = _.shuffle(useMixtapeProjects());
    const me = useMySession();
    const myProjects = useMixtapeProjects().filter(p => p.Owners.includes(me.DiscordID));

    return <Container>
        <Row className={"row-cols-1"}>
            <Col>
                <h1 className={"title"} style={{textAlign: "left"}}>Wintry Mix | Members Dashboard</h1>
            </Col>
            <Col className={"mt-3"}>
                <Row md={2} sm={1}>
                    {projects.map((p, i) =>
                        <Col key={i.toString()} className={"mt-3"}>
                            <ProjectCard me={me} project={p}/>
                        </Col>)}
                </Row>
            </Col>
        </Row>
    </Container>;
};

const ProjectCard = (props: { project: MixtapeProject, me: Session }) => {
    const {project, me} = props;

    const [showEdit, setShowEdit] = useState(false);
    const blurbRef = useRef({} as HTMLTextAreaElement);
    const tagsRef = useRef({} as HTMLInputElement);

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
                        defaultValue={project.Blurb}/>
                </InputGroup> :
                <ReactMarkdown>
                    {project.Blurb}
                </ReactMarkdown>}
            <Row>
                <Col>
                    {showEdit ?
                        <InputGroup className="mb-3">
                            <InputGroup.Text children={"#"}/>
                            <FormControl
                                ref={tagsRef}
                                defaultValue={project.Tags.join(", ")}/>
                        </InputGroup> :
                        <Card.Text>
                            <i># {project.Tags.join(", ")}</i>
                        </Card.Text>}
                </Col>
                <Col className={"d-flex justify-content-end"}>
                    {showEdit ?
                        <Button
                            type={"button"}
                            variant={"outline-secondary"}
                            size={"sm"}
                            onClick={buttonOnClick}
                            children={"Submit"}/> :
                        <Button
                            type={"button"}
                            variant={"outline-secondary"}
                            size={"sm"}
                            onClick={() => setShowEdit(true)}
                            children={"Edit"}/>}
                </Col>
            </Row>
        </Card.Body>
    </Card>;
};
