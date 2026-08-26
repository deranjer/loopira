import { Button, Paper, Stack, Text, TextInput, PasswordInput, Title } from '@mantine/core'
import { useForm } from '@mantine/form'
import { useNavigate } from 'react-router-dom'
import { ApiError } from '../lib/api/client'
import { useCompleteSetup } from '../lib/api/hooks'

export function Setup() {
  const navigate = useNavigate()
  const complete = useCompleteSetup()
  const form = useForm({
    initialValues: { name: '', email: '', password: '', confirmPassword: '' },
    validate: {
      confirmPassword: (value, values) => (value !== values.password ? 'Passwords do not match' : null),
      password: (value) => (value.length < 8 ? 'Password must be at least 8 characters' : null),
    },
  })

  function submit(values: typeof form.values) {
    complete.mutate(
      { name: values.name, email: values.email, password: values.password },
      { onSuccess: () => navigate('/issues') },
    )
  }

  const error =
    complete.error instanceof ApiError ? complete.error.message : complete.isError ? 'Setup failed' : null

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'var(--mantine-color-dark-7)',
      }}
    >
      <Paper w={420} p="xl" radius="md" bg="dark.6" style={{ border: '1px solid #1d1e21' }}>
        <form onSubmit={form.onSubmit(submit)}>
          <Stack gap="md">
            <div>
              <img src="/favicon.svg" width={34} height={34} alt="Loopira" style={{ display: 'block', marginBottom: 14 }} />
              <Title order={2} c="dark.0">
                Welcome to Loopira
              </Title>
              <Text size="sm" c="dimmed" mt={4}>
                Create the admin account to finish setting up your workspace.
              </Text>
            </div>
            <TextInput label="Name" placeholder="Ada Lovelace" required {...form.getInputProps('name')} />
            <TextInput label="Email" placeholder="you@company.com" required {...form.getInputProps('email')} />
            <PasswordInput label="Password" required {...form.getInputProps('password')} />
            <PasswordInput label="Confirm password" required {...form.getInputProps('confirmPassword')} />
            {error && (
              <Text c="red" size="sm">
                {error}
              </Text>
            )}
            <Button type="submit" fullWidth loading={complete.isPending}>
              Create admin account
            </Button>
          </Stack>
        </form>
      </Paper>
    </div>
  )
}
