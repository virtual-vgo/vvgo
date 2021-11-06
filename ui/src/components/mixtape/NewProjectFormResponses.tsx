import _ from "lodash"
import {Table} from "react-bootstrap";
import {useSheet} from "../../datasets";
import {InternalOopsie} from "../errors/InternalOopsie";
import {LoadingText} from "../shared/LoadingText";
import {RootContainer} from "../shared/RootContainer";

export const NewProjectFormResponses = () => {
    const sheet = useSheet("wintry_mix_form_responses", "Form Responses 1");
    if (!sheet) return <RootContainer><LoadingText/></RootContainer>;
    if (_.isEmpty(sheet.Values)) return <InternalOopsie/>;
    return <RootContainer title="New Project Form Responses">
        <Table>
            <thead>
            <tr>
                {sheet.Values[0].map(v => <td key={v}>{v}</td>)}
            </tr>
            </thead>
            <tbody>
            {sheet.Values.slice(1).map(r =>
                <tr key={r.join(",")}>
                    {r.map(v => <td key={v}>{v}</td>)}
                </tr>)}
            </tbody>
        </Table>
    </RootContainer>;
};

export default NewProjectFormResponses
