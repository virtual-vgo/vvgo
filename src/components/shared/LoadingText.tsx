import { Box, CircularProgress } from "@mui/material";

export const LoadingText = () => {
    return (
        <Box display="flex" justifyContent="center" alignItems="center" padding={8}>
            <CircularProgress />
        </Box>
    );
};
