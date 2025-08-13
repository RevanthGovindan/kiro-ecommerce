import { ErrorPage } from '@/components/error/ErrorPage';

export default function NotFound() {
  return (
    <ErrorPage
      statusCode={404}
      title="Page Not Found"
      message="Sorry, we couldn't find the page you're looking for. It might have been moved, deleted, or you entered the wrong URL."
    />
  );
}