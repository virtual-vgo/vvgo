import {useState} from "react";
import {GuildMember, useGuildMembers} from "../datasets";
import {Container} from "../components";
import {Button, Col, Form, ListGroup, Row} from "react-bootstrap";
import React = require("react");


export const NewProjectWorkflow = () => {
    const limit = 5
    const [name, setName] = useState("")
    const [query, setQuery] = useState("")
    const [owners, setOwners] = useState([] as GuildMember[])
    const guildMembers = useGuildMembers(query, limit)

    return <Container>
        <h1>Winter Mixtape</h1>
        <h2>New Project Workflow</h2>
        <Row>
            <Col className={'col-sm-4'}>
                <CreateProjectForm
                    name={name}
                    setName={setName}
                    owners={owners}
                    setOwners={setOwners}
                    setQuery={setQuery}
                    guildMembers={guildMembers}
                />
            </Col>
        </Row>
    </Container>
}

const CreateProjectForm = (props: {
    name: string
    guildMembers: GuildMember[]
    owners: GuildMember[]
    setName: (x: string) => void
    setQuery: (x: string) => void
    setOwners: (x: GuildMember[]) => void
}) => {
    const handleNameChange = ({}) => {
        const value = ""
        props.setName(value)
    }
    const handleQueryChange = ({}) => {
        const value = ""
        props.setQuery(value ? value : "")
    }
    const handleOwnerChange = (member: GuildMember) => () => {
        props.setOwners([...props.owners, member])
    }

    const handleSubmit = () =>
        console.log(`name: ${props.name}, owners: ${props.owners.map(o => o.nick).join(", ")}`)

    return <Form>
        <Form.Group className="mb-3" controlId="inputTrackName">
            <Form.Label>Track Name</Form.Label>
            <Form.Control type="text" placeholder="tRaCK naMe" onChange={handleNameChange}/>
        </Form.Group>
        <Form.Group className="mb-3" controlId="inputOwners">
            <Form.Label>Owners</Form.Label>
            <Form.Control type="text" placeholder="search members" onChange={handleQueryChange}/>
        </Form.Group>

        <div className="mb-3">
            {[...props.guildMembers.filter(m => m.nick && m.nick !== ""), ...props.owners]
                .map(m => <Form.Check
                    inline
                    label={m.nick}
                    type="checkbox"
                    onChange={handleOwnerChange(m)}
                />)}
        </div>

        <ListGroup className="mb-3">

        </ListGroup>
        <Button variant="primary" type="button" onClick={handleSubmit}>
            Submit
        </Button>
    </Form>
}
