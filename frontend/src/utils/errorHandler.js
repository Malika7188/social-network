import { showToast } from "@/components/ui/ToastContainer";

export const handleApiError = async (
  error,
  defaultMessage = "An error occurred"
) => {
  console.error(error);

  let errorMessage = defaultMessage;

  try {
    if (error.response) {
      const errorData = await error.response.json();
      errorMessage = errorData.error || defaultMessage;
    }
  } catch (e) {
    console.error("Error parsing error response:", e);
  }

  showToast(errorMessage, "error");
  return errorMessage;
};
