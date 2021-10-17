import React, {useState} from "react";
import {Render} from "./render";
import {deleteSessions, Session, SessionKinds, useSessions} from "./models";
import _ from "lodash";

export const Entrypoint = (selectors) => Render(<Sessions/>, selectors)

const Sessions = () => {
    const sessions = useSessions().Sessions
    const [refresh, setRefresh] = useState(0)
    const [deleteButtonState, setDeleteButtonState] = useState(new Map())
    sessions.sort((a, b) => a.Expires - b.Expires)
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
                    <NewSession buttonState={deleteButtonState} setButtonState={setDeleteButtonState}/>
                    {sessions
                        .filter(session => deleteButtonState.get(session.Key) !== 'deleted')
                        .map((x, i) =>
                            <SessionRow key={i} session={x} buttonState={deleteButtonState}
                                        setButtonState={setDeleteButtonState}/>)}
                    </tbody>
                </table>
            </div>
        </div>
    </div>
}

const NewSession = (props) => {
    const [session, setSession] = React.useState(new Session())
    session.Key = "new"

    return <tr>
        <td>
            <select className="custom-select mr-sm-2" id="selectKind">
                <option defaultValue>Kind...</option>
                {_.keys(SessionKinds).map(k => <option key={k} value={k}>{k}</option>)}
            </select>
        </td>
        <td>
            <select className="custom-select mr-sm-2" id="selectRoles">
                <option defaultValue>Roles...</option>
                {['write_spreadsheet'].map(k => <option key={k} value={k}>{k}</option>)}
            </select>
        </td>
        <td/>
        <td><input type={'number'} className={'form-control'} id={'selectExpires'}/></td>
        <td width={120}>
            <DeleteButton session={session} buttonState={props.buttonState} setButtonState={props.setButtonState}/>
        </td>
    </tr>
}

const CreateButton = (props) => {
    const buttonState = props.buttonState
    const setButtonState = props.setButtonState

    const buttonClick = (event) => {
        setButtonState('creating')
        new Promise(resolve => setTimeout(resolve, 500)
        ).then(() => console.log(event)
        ).then(() => setButtonState('created')
        ).catch(error => console.log(error))
    }

    if (buttonState === 'ready')
        return <button className={'btn btn-sm btn-dark btn-outline-dark text-light w-100'}
                       onClick={() => buttonClick(props.session)}>Create</button>
    if (buttonState === 'created')
        return <button className={'btn btn-sm btn-warning text-warning w-100'}>☠️☠️☠️</button>
    return <div/>
}

const SessionRow = (props) => {
    return <tr>
        <td>{props.session.Kind}</td>
        <td>{props.session.Roles.reduce((a, b) => a + ", " + b)}</td>
        <td>{props.session.DiscordID}</td>
        <td>{props.session.Expires}</td>
        <td width={120}>
            <DeleteButton session={props.session} buttonState={props.buttonState}
                          setButtonState={props.setButtonState}/>
        </td>
    </tr>
}

const DeleteButton = (props) => {
    const buttonState = props.buttonState
    const setButtonState = props.setButtonState

    const buttonClick = (session, event) => {
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
    const status = (key) => {
        if (key === 'new') return 'new'
        if (buttonState.has(key)) return buttonState.get(key)
        return 'ready'
    }

    if (status(key) === 'ready')
        return <button className={'btn btn-sm btn-dark btn-outline-dark text-light w-100'}
                       onClick={() => buttonClick(props.session)}>Delete</button>
    if (status(key) === 'deleting')
        return <button className={'btn btn-sm btn-warning text-warning w-100'}>☠️☠️☠️</button>
    return <div/>
}
