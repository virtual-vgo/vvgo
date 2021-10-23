import React = require("react");
import {Container} from "../components";
import {Col, Row} from "react-bootstrap";
import {useMixtapeProjects, useMySession} from "../datasets";

export const HostDashboard = () => {
    const me = useMySession()
    const myProjects = useMixtapeProjects().filter(p => p.Owners.includes(me.DiscordID))
    return <Container>
        <Row>
            {myProjects.map(p =>
                <Col><MixtapeProject key={p.Name}/></Col>)}
        </Row>
    </Container>
}

const MixtapeProject = () => {
    return <div/>
}
