import React, {useState} from "react";
import {Render} from "./render";
import {createSessions, deleteSessions, SessionKinds, useMySession, useSessions} from "./models";
import _ from "lodash";

export const Entrypoint = (selectors) => Render(<Sessions/>, selectors)

const Sessions = () => {
    const [me, _] = useMySession()
    const [sessions, setSessions] = useSessions()
    const [deleteButtonState, setDeleteButtonState] = useState(new Map())
    const [createButtonState, setCreateButtonState] = useState('new')
    sessions.sort((a, b) => a.ExpiresAt - b.ExpiresAt)

    const MySessions = () => {
        return sessions
            .filter(session => deleteButtonState.get(session.Key) !== 'deleted')
            .filter(session => session.DiscordID === me.DiscordID)
            .map(session => <SessionRow key={session.Key} session={session} buttonState={deleteButtonState}
                                        setButtonState={setDeleteButtonState}/>)
    }

    const OtherSessions = () => {
        return sessions
            .filter(session => deleteButtonState.get(session.Key) !== 'deleted')
            .filter(session => session.DiscordID !== me.DiscordID)
            .map(session => <SessionRow className={'text-warning'} key={session.Key} session={session}
                                        buttonState={deleteButtonState}
                                        setButtonState={setDeleteButtonState}/>)
    }

    return <div className={'container mt-4'}>
        <div className={'row row-cols-1 mt-2'}>
            <div className={'col'}>
                <h1>Sessions</h1>
                <table className={'table text-light'}>
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
                    <NewSession buttonState={createButtonState} setButtonState={setCreateButtonState}
                                sessions={sessions} setSessions={setSessions}/>
                    <MySessions/>
                    <OtherSessions/>
                    </tbody>
                </table>
            </div>
        </div>
    </div>
}

const NewSession = (props) => {
    const inputKind = React.useRef()
    const inputRoles = React.useRef()
    const inputExpires = React.useRef()

    const roles = ['write_spreadsheet']
    return <tr>
        <td>
            <select className="custom-select mr-sm-2" ref={inputKind}>
                <option defaultValue>{SessionKinds.ApiToken}</option>
                {_.keys(SessionKinds)
                    .filter(k => k !== 'ApiToken')
                    .map(k => <option key={k} value={SessionKinds[k]}>{SessionKinds[k]}</option>)}
            </select>
        </td>
        <td>
            <select className="custom-select mr-sm-2" ref={inputRoles}>
                <option defaultValue>{roles[0]}</option>
                {roles.slice(1, -1).map(k => <option key={k} value={k}>{k}</option>)}
            </select>
        </td>
        <td/>
        <td><input type={'number'} className={'form-control'} ref={inputExpires} defaultValue={3600}/></td>
        <td width={120}>
            <CreateButton sessions={props.sessions} setSessions={props.setSessions}
                          inputKind={inputKind} inputRoles={inputRoles} inputExpires={inputExpires}
                          buttonState={props.buttonState} setButtonState={props.setButtonState}/>
        </td>
    </tr>
}

const CreateButton = (props) => {
    const buttonClick = (e) => {
        props.setButtonState('creating')
        new Promise(resolve => setTimeout(resolve, 500)
        ).then(() => createSessions([{
                kind: props.inputKind.current.value,
                roles: [props.inputRoles.current.value],
                expires: Number(props.inputExpires.current.value)
            }])
        ).then(resp => {
            console.log("Created sessions:", resp)
            props.setSessions([...resp, ...props.sessions])
            props.setButtonState('created')
        }).catch(error => console.log(error))
    }

    if (props.buttonState !== 'creating')
        return <button className={'btn btn-sm btn-dark btn-outline-dark text-light w-100'}
                       onClick={buttonClick}>
            Create
        </button>
    return <button className={'btn btn-sm btn-dark btn-outline-dark text-light w-100'}>
        Creating
    </button>
}

const SessionRow = (props) => {
    const session = props.session
    const expiresAt = () => {
        if (session.ExpiresAt) {
            const expiresAt = session.ExpiresAt
            const date = expiresAt.toDateString()
            const time = `${expiresAt.toLocaleTimeString()}`
            return expiresAt.toLocaleString()
        }
        return ""
    }
    return <tr className={props.className}>
        <td>{session.Kind}</td>
        <td>{session.Roles.reduce((a, b) => a + ", " + b)}</td>
        <td>{session.DiscordID}</td>
        <td>{expiresAt()}</td>
        <td width={120}>
            <DeleteButton session={props.session} buttonState={props.buttonState}
                          setButtonState={props.setButtonState}/>
        </td>
    </tr>
}

const DeleteButton = (props) => {
    const buttonState = props.buttonState
    const setButtonState = props.setButtonState

    const buttonClick = (e) => {
        const session = props.session
        const newState = new Map()
        buttonState.forEach((val, key) => newState.set(key, val))
        newState.set(session.Key, 'deleting')
        setButtonState(newState)

        new Promise(resolve => setTimeout(resolve, 500)
        ).then(() => deleteSessions([session.Key])
        ).then(() => {
            const state = new Map()
            buttonState.forEach((val, key) => state.set(key, val))
            state.set(session.Key, 'deleted')
            setButtonState(state)
        }).catch(error => console.log(error))
    }

    const key = props.session.Key

    if (buttonState.get(key) === 'deleted')
        return <div/>
    if (buttonState.get(key) === 'deleting')
        return <button className={'btn btn-sm btn-warning text-warning w-100'}>☠️☠️☠️</button>

    return <button className={'btn btn-sm btn-dark btn-outline-dark text-light w-100'}
                   onClick={buttonClick}>Delete</button>
}
