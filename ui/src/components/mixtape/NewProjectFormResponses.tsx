import {isEmpty} from "lodash/fp";
import {Table} from "react-bootstrap";
import {useSheet} from "../../datasets";
import {InternalOopsie} from "../errors/InternalOopsie";
import {LoadingText} from "../shared/LoadingText";

export const NewProjectFormResponses = () => {
    const sheet = useSheet("wintry_mix_form_responses", "Form Responses 1");
    if (!sheet) return <LoadingText/>;
    if (isEmpty(sheet.Values)) return <InternalOopsie/>;
    return <div>
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
    </div>;
};

export default NewProjectFormResponses;
