import React = require("react");
import Button from "react-bootstrap/Button";
import ButtonGroup from "react-bootstrap/ButtonGroup";
import Col from "react-bootstrap/Col";
import Form from "react-bootstrap/Form";
import Row from "react-bootstrap/Row";
import {Redirect} from "react-router";
import {oauthRedirect, passwordLogin} from "../../auth";
import {RootContainer} from "../shared/RootContainer";

const styles = {
    Form: {
        width: "100%",
        maxWidth: "330px",
        padding: "15px",
        margin: "auto",
    } as React.CSSProperties,
};

export const RedirectLogin = () => <Redirect to="/login/"/>
export const RedirectLoginSuccess = () => <Redirect to="/parts"/>;
export const RedirectLoginFailure = () => <Redirect to="/login/failure"/>;

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
    return <RootContainer title="Login">
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
