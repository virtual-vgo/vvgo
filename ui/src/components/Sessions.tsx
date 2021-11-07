import {Dispatch, MutableRefObject, SetStateAction, useRef, useState} from "react";
import {Button} from "react-bootstrap";
import {getSession} from "../auth";
import {createSessions, deleteSessions, Session, SessionKind, useSessions} from "../datasets";
import {LoadingText} from "./shared/LoadingText";

export const Sessions = () => {
    const me = getSession();
    const [sessions, setSessions] = useSessions();
    const [deleteButtonState, setDeleteButtonState] = useState(new Map());
    const [createButtonState, setCreateButtonState] = useState("new");

    if (!sessions) return <LoadingText/>;
    sessions.sort((a, b) => new Date(a.ExpiresAt ?? "").getTime() - new Date(b.ExpiresAt ?? "").getTime());

    const mySessions = sessions
        .filter(session => deleteButtonState.get(session.Key) !== "deleted")
        .filter(session => session.DiscordID === me.DiscordID)
        .map(session => <SessionRow
            key={session.Key}
            session={session}
            buttonState={deleteButtonState}
            setButtonState={setDeleteButtonState}
        />);

    const otherSessions = sessions
        .filter(session => deleteButtonState.get(session.Key) !== "deleted")
        .filter(session => session.DiscordID !== me.DiscordID)
        .map(session => <SessionRow
            className={"text-warning"}
            key={session.Key}
            session={session}
            buttonState={deleteButtonState}
            setButtonState={setDeleteButtonState}
        />);

    return <div>
        <div className={"row row-cols-1 mt-2"}>
            <div className={"col"}>
                <h1>Sessions</h1>
                <table className={"table text-light"}>
                    <thead>
                    <tr>
                        <th>Kind</th>
                        <th>Roles</th>
                        <th>Discord ID</th>
                        <th>Expires</th>
                        <th/>
                    </tr>
                    </thead>
                    <tbody>
                    <NewSession
                        buttonState={createButtonState}
                        setButtonState={setCreateButtonState}
                        sessions={sessions}
                        setSessions={setSessions}/>
                    {mySessions}
                    {otherSessions}
                    </tbody>
                </table>
            </div>
        </div>
    </div>;
};

const NewSession = (props: {
    sessions: Session[];
    setSessions: (sessions: Session[]) => void;
    buttonState: string;
    setButtonState: Dispatch<SetStateAction<string>>;
}): JSX.Element => {
    const inputKind = useRef({} as HTMLSelectElement);
    const inputRoles = useRef({} as HTMLSelectElement);
    const inputExpires = useRef({} as HTMLInputElement);

    const roles = ["write_spreadsheet"];
    return <tr>
        <td>
            <select className="custom-select mr-sm-2" ref={inputKind}>
                <option>{SessionKind.ApiToken}</option>
                {}
                {Object.entries(SessionKind)
                    .filter(([k]) => k !== "ApiToken")
                    .map(([k, v]) => <option key={k} value={v}>{v}</option>)}
            </select>
        </td>
        <td>
            <select className="custom-select mr-sm-2" ref={inputRoles}>
                <option>{roles[0]}</option>
                {roles.slice(1, -1).map(k => <option key={k} value={k}>{k}</option>)}
            </select>
        </td>
        <td/>
        <td>{/*CreatedAt*/}</td>
        <td><input type={"number"} className={"form-control"} ref={inputExpires} defaultValue={3600}/></td>
        <td width={120}>
            <CreateButton sessions={props.sessions} setSessions={props.setSessions}
                          inputKind={inputKind} inputRoles={inputRoles} inputExpires={inputExpires}
                          buttonState={props.buttonState} setButtonState={props.setButtonState}/>
        </td>
    </tr>;
};

const CreateButton = (props: {
    setButtonState: Dispatch<SetStateAction<string>>;
    inputKind: MutableRefObject<HTMLSelectElement>;
    inputRoles: MutableRefObject<HTMLSelectElement>;
    inputExpires: MutableRefObject<HTMLInputElement>;
    setSessions: (sessions: Session[]) => void;
    sessions: Session[];
    buttonState: string;
}) => {
    const buttonClick = () => {
        props.setButtonState("creating");
        createSessions([{
            Kind: props.inputKind.current.value as SessionKind,
            Roles: [props.inputRoles.current.value] as string[],
            expires: Number(props.inputExpires.current.value),
        }])
            .then(resp => {
                console.log("Created sessions:", resp);
                props.setSessions([...resp, ...props.sessions]);
                props.setButtonState("created");
            })
            .catch(error => console.log(error));
    };

    if (props.buttonState !== "creating")
        return <button className={"btn btn-sm btn-dark btn-outline-dark text-light w-100"}
                       onClick={buttonClick}>
            Create
        </button>;
    return <button className={"btn btn-sm btn-dark btn-outline-dark text-light w-100"}>
        Creating
    </button>;
};

const SessionRow = (props: {
    session: Session;
    buttonState: Map<string, string>;
    className?: string;
    setButtonState: Dispatch<SetStateAction<Map<string, string>>>;
}) => {
    const session = props.session;
    return <tr className={props.className}>
        <td>
            <Button
                variant={"outline-primary"}
                className={"borderless"}
                onClick={() => navigator.clipboard.writeText(session.Key ?? "")}>
                {session.Kind}
            </Button>
        </td>
        <td>{session.Roles ? session.Roles.join(", ") : "none"}</td>
        <td>{session.DiscordID}</td>
        <td>{new Date(session.CreatedAt ?? "").toLocaleString()}</td>
        <td>{new Date(session.ExpiresAt ?? "").toLocaleString()}</td>
        <td width={120}>
            <DeleteButton
                session={props.session}
                buttonState={props.buttonState}
                setButtonState={props.setButtonState}/>
        </td>
    </tr>;
};

const DeleteButton = (props: {
    buttonState: Map<string, string>;
    setButtonState: Dispatch<SetStateAction<Map<string, string>>>;
    session: Session;
}) => {
    const buttonState = props.buttonState;
    const setButtonState = props.setButtonState;

    const buttonClick = () => {
        const session = props.session;
        const newState = new Map();
        buttonState.forEach((val, key) => newState.set(key, val));
        newState.set(session.Key, "deleting");
        setButtonState(newState);

        deleteSessions([session.Key ?? ""]).then(() => {
            const state = new Map();
            buttonState.forEach((val, key) => state.set(key, val));
            state.set(session.Key, "deleted");
            setButtonState(state);
        }).catch(error => console.log(error));
    };

    const sessionKey = props.session.Key ?? "";
    if (buttonState.get(sessionKey) === "deleted")
        return <div/>;
    if (buttonState.get(sessionKey) === "deleting")
        return <button className={"btn btn-sm btn-warning text-warning w-100"}>☠️☠️☠️</button>;

    return <button className={"btn btn-sm btn-dark btn-outline-dark text-light w-100"}
                   onClick={buttonClick}>Delete</button>;
};
export default Sessions;
