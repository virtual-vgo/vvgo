import imgSrc from "./401.gif";
import { ErrorPage } from "./ErrorPage";

export const AccessDenied = () => (
  <ErrorPage src={imgSrc} alt="401 Access Denied" />
);
export default AccessDenied;
