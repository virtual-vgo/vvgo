import _ = require("lodash");
import React = require("react");
import Button from "react-bootstrap/Button";
import ButtonGroup from "react-bootstrap/ButtonGroup";
import Col from "react-bootstrap/Col";
import Form from "react-bootstrap/Form";
import Row from "react-bootstrap/Row";
import {Redirect} from "react-router";
import {Link} from "react-router-dom";
import {discordLogin, oauthRedirect, passwordLogin} from "../auth";
import {GeekSquad} from "./shared/GeekSquad";
import {RootContainer} from "./shared/RootContainer";

const styles = {
    Form: {
        width: "100%",
        maxWidth: "330px",
        padding: "15px",
        margin: "auto",
    } as React.CSSProperties,
};

export const RedirectLoginSuccess = () => <Redirect to="/login/success"/>;
export const RedirectLoginFailure = () => <Redirect to="/login/failure"/>;

export const LoginSuccess = () => {
    return <RootContainer>
        <div className="row-cols-1">
            <h1>Success!</h1>
            <h2><a href="/login/redirect" className="btn btn-link text-light">Continue to parts...</a></h2>
        </div>
    </RootContainer>;
};

export const LoginDiscord = () => {
    const [success, setSuccess] = React.useState(false);
    const [failed, setFailed] = React.useState(false);

    const params = new URLSearchParams(window.location.search);
    const code = _.defaultTo(params.get("code"), "");
    const state = _.defaultTo(params.get("state"), "");

    React.useEffect(() => {
        discordLogin(code, state)
            .then(me => {
                setSuccess(true);
                console.log("login successful", me);
            })
            .catch(err => {
                setFailed(true);
                console.log("login failed", err);
            });
    });

    switch (true) {
        case success:
            return <RedirectLoginSuccess/>;
        case failed:
            return <RedirectLoginFailure/>;
        default:
            return <div>Loading...</div>;
    }
};

export const LoginFailure = () => {
    return <p>
        Please join our <a href="https://discord.gg/vvgo">Discord server</a> and accept the rules before logging in with
        Discord.
        <br/>
        If you think you should be able to login, please check <GeekSquad/>.
        <br/>
        <br/>
        <Link to="/login">Return to the login page.</Link>
    </p>;
};

export const Login = () => {
    const [success, setSuccess] = React.useState(false);
    const [loginFailed, setLoginFailed] = React.useState(false);
    const userRef = React.useRef({} as HTMLInputElement);
    const passRef = React.useRef({} as HTMLInputElement);

    const onClickLogin = () =>
        passwordLogin(userRef.current.value, passRef.current.value)
            .then(me => {
                setSuccess(true);
                console.log("login successful", me);
            })
            .catch(err => {
                setLoginFailed(true);
                console.log("login failed", err);
            });

    const onClickDiscordLogin = () =>
        oauthRedirect()
            .then((data: { DiscordURL: string; }) => {
                document.location.href = data.DiscordURL;
            })
            .catch((err: unknown) => {
                console.log("api error", err);
            });

    if (success) return <RedirectLoginSuccess/>;

    // noinspection HtmlUnknownTarget
    return <RootContainer>
        <Row className="justify-content-md-center">
            <Col style={styles.Form}>
                <Form>
                    <img className="mb-4 mt-4" src="/images/logo.svg" alt="" width="100%"/>
                    <Form.Group>
                        <Form.Control ref={userRef} type="text" placeholder="user" required autoFocus/>
                        <Form.Control ref={passRef} type="password" placeholder="password" required/>
                    </Form.Group>
                    <div hidden={!loginFailed} className="mt-2 text-warning">Invalid user or password 😔</div>
                </Form>
                <div className={"d-grid " + (loginFailed ? "mt-2" : "mt-4")}>
                    <ButtonGroup vertical>
                        <Button
                            size="lg"
                            className="bg-vvgo-purple text-light"
                            type="button"
                            onClick={onClickLogin}>
                            Sign in
                        </Button>
                        <Button
                            size="lg"
                            type="button"
                            className="bg-discord-blue text-light"
                            onClick={onClickDiscordLogin}>
                            Sign in with Discord
                        </Button>
                    </ButtonGroup>
                </div>
            </Col>
        </Row>
    </RootContainer>;
};
