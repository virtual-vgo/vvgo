import React = require("react");
import Button from "react-bootstrap/Button";
import ButtonGroup from "react-bootstrap/ButtonGroup";

export const ShowHideToggle = (props: { title: string, state: boolean, setState: (x: boolean) => void }) => {
    return <div className={"m-2"}>
        <strong>{props.title}</strong><br/>
        <ButtonGroup>
            <Button
                size={"sm"}
                className={"text-light"}
                onClick={() => props.setState(true)}
                variant={props.state ? "warning" : ""}>
                Show
            </Button>
            <Button
                size={"sm"}
                className={"text-light"}
                onClick={() => props.setState(false)}
                variant={props.state ? "" : "primary"}>
                Hide
            </Button>
        </ButtonGroup>
    </div>;
};
