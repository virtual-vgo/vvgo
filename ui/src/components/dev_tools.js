import React from "react";
import {Status} from "./hooks";
import FormControlLabel from "@material-ui/core/FormControlLabel";
import Checkbox from "@material-ui/core/Checkbox";
import FormGroup from "@material-ui/core/FormGroup";
import CheckBoxOutlineBlankIcon from '@material-ui/icons/CheckBoxOutlineBlank';

export default function DevTools(props) {
    if (props.apiRoles.data.includes("vvgo-teams") === false) {
        return null
    }

    return <div>
        <ChooseRolesForm {...props}/>
    </div>
}

function ChooseRolesForm(props) {
    if (props.apiRoles.status !== Status.Complete) {
        return null
    }

    let roleChoices = [...props.apiRoles.data, 'anonymous']
    roleChoices = roleChoices.filter((e, i) => roleChoices.indexOf(e) === i)

    const roleSelection = {}
    props.uiRoles.data.forEach(role => roleSelection[role] = true)

    const handleChange = (event) => {
        roleSelection[event.target.name] = event.target.checked
        const newRoles = Object.keys(roleSelection).filter(key => roleSelection[key])
        console.log("new ui roles", newRoles)
        props.uiRoles.setData(newRoles)
    }

    function RoleCheckbox(props) {
        const checkbox = <Checkbox color='primary' name={props.name} checked={props.checked} onChange={handleChange}
                                   icon={<CheckBoxOutlineBlankIcon color='primary'/>}/>
        return <FormControlLabel
            control={checkbox}
            label={props.name}
        />
    }

    return <FormGroup row>
        {roleChoices.map(role => <RoleCheckbox key={role} name={role} checked={roleSelection[role]}/>)}
    </FormGroup>
}
