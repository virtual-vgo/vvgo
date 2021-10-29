import React = require("react");
import Col from "react-bootstrap/Col";
import Row from "react-bootstrap/Row";

export const Footer = () => <Row className="text-center m-4">
    <Col>
        <a className="text-light"
           href="https://www.youtube.com/channel/UCeipEtsfjAA_8ATsd7SNAaQ">
            <i className="fab fa-youtube fa-2x"/>
        </a>
        <a className="text-light" href="https://www.facebook.com/groups/1080154885682377/">
            <i className="fab fa-facebook fa-2x"/>
        </a>
        <a className="text-light"
           href="https://vvgo.bandcamp.com/">
            <i className="fab fa-bandcamp fa-2x"/>
        </a>
        <a className="text-light" href="https://github.com/virtual-vgo/vvgo">
            <i className="fab fa-github fa-2x"/>
        </a>
        <a className="text-light"
           href="https://www.instagram.com/virtualvgo/">
            <i className="fab fa-instagram fa-2x"/>
        </a>
        <a className="text-light" href="https://twitter.com/virtualvgo">
            <i className="fab fa-twitter fa-2x"/>
        </a>
        <a className="text-light" href="https://discord.gg/vvgo">
            <i className="fab fa-discord fa-2x"/>
        </a>
    </Col>
</Row>;