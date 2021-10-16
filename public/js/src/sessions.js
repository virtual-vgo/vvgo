import React, {useState} from "react";
import {Render} from "./render";
import {deleteSessions, useSessions} from "./models";

export const Entrypoint = (selectors) => Render(<Sessions/>, selectors)

const Sessions = () => {
    const sessions = useSessions()

    const [buttonState, setButtonState] = useState(new Map())
    return <div className={'container'}>
        <table className={'table text-light'}>
            <thead>
            <tr>
                <th>Kind</th>
                <th>Roles</th>
                <th>Discord ID</th>
                <th/>
            </tr>
            </thead>
            <tbody>
            {sessions
                .filter(session => buttonState.get(session.Key) !== 'gone')
                .map((x, i) =>
                    <SessionRow key={i} session={x} buttonState={buttonState} setButtonState={setButtonState}/>)}
            </tbody>
        </table>
    </div>
}

const SessionRow = (props) => {
    return <tr>
        <td>{props.session.Kind}</td>
        <td>{props.session.Roles.reduce((a, b) => a + ", " + b)}</td>
        <td>{props.session.DiscordID}</td>
        <td>
            <Button session={props.session} buttonState={props.buttonState} setButtonState={props.setButtonState}/>
        </td>
    </tr>
}

const Button = (props) => {
    const buttonState = props.buttonState
    const setButtonState = props.setButtonState

    const buttonClick = (session) => {
        const newState = new Map()
        buttonState.forEach((val, key) => newState.set(key, val))
        newState.set(session.Key, 'deleting')
        setButtonState(newState)
        new Promise(resolve => setTimeout(resolve, 500)
        ).then(() =>
            deleteSessions([session.Key])
        ).then(() => {
            const state = new Map()
            buttonState.forEach((val, key) => state.set(key, val))
            state.set(session.Key, 'deleted')
            setButtonState(state)
        }).then(resolve =>
            setTimeout(resolve, 500)
        ).then(() => {
            const state = new Map()
            buttonState.forEach((val, key) => state.set(key, val))
            state.set(session.Key, 'gone')
            setButtonState(state)
        }).catch(error => console.log(error))
    }

    const key = props.session.Key
    const status = buttonState.has(key) ? buttonState.get(key) : 'ready'
    switch (status) {
        case 'ready':
            return <button className={'btn btn-sm btn-dark btn-outline-dark text-light'}
                           onClick={() => buttonClick(props.session)}>Delete</button>
        case 'deleting':
            return <p className={'text-warning'}>Deleting...</p>
        case 'deleted':
            return <p className={'text-light'}>☠️☠️☠️</p>
        default:
            return <div/>
    }
}
