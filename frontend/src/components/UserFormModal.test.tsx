import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach } from "vitest";
import { UserFormModal } from "./UserFormModal";

// 1. Mock do hook usePermission
const mockUsePermission = vi.fn();
vi.mock("../hooks/usePermission", () => ({
  usePermission: (perm: string) => mockUsePermission(perm),
}));

// 2. Mock do react-query's useQuery
vi.mock("@tanstack/react-query", () => ({
  useQuery: () => ({
    data: [
      { id: 1, name: "Admin" },
      { id: 2, name: "Operator" },
    ],
    isLoading: false,
  }),
}));

// 3. Mock do api.ts
vi.mock("../services/api", () => ({
  api: {
    roles: {
      list: vi.fn(),
    },
  },
}));

describe("UserFormModal", () => {
  const defaultProps = {
    user: null,
    pending: false,
    error: null,
    onClose: vi.fn(),
    onSubmit: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockUsePermission.mockReturnValue(true); // Permissão padrão concedida
  });

  it("renders 'Criar Novo Usuário' in creation mode", () => {
    render(<UserFormModal {...defaultProps} />);
    expect(screen.getByText("Criar Novo Usuário")).toBeInTheDocument();
    expect(screen.getByPlaceholderText("Nome do usuário")).toBeInTheDocument();
  });

  it("renders 'Editar Usuário' in edit mode", () => {
    const editProps = {
      ...defaultProps,
      user: {
        id: 1,
        name: "João Silva",
        email: "joao@empresa.com",
        active: true,
        roleId: 2,
        role: { id: 2, name: "Operator" },
        permissions: [],
      },
    };
    render(<UserFormModal {...editProps} />);
    expect(screen.getByText("Editar Usuário")).toBeInTheDocument();
    expect(screen.getByDisplayValue("João Silva")).toBeInTheDocument();
    expect(screen.getByDisplayValue("joao@empresa.com")).toBeInTheDocument();
  });

  it("shows error if password is less than 12 characters", async () => {
    render(<UserFormModal {...defaultProps} />);
    
    // Preencher campos obrigatórios
    fireEvent.change(screen.getByPlaceholderText("Nome do usuário"), { target: { value: "Fulano de Tal" } });
    fireEvent.change(screen.getByPlaceholderText("exemplo@empresa.com"), { target: { value: "fulano@empresa.com" } });
    
    // Senha muito curta
    fireEvent.change(screen.getByPlaceholderText("Senha (mínimo de 12 caracteres)"), { target: { value: "Senha123" } });
    
    // Submeter
    fireEvent.submit(screen.getByRole("button", { name: "Criar usuário" }));

    expect(screen.getByText("A senha deve possuir pelo menos 12 caracteres.")).toBeInTheDocument();
    expect(defaultProps.onSubmit).not.toHaveBeenCalled();
  });

  it("shows error if password has no upper, lower, or digit", () => {
    render(<UserFormModal {...defaultProps} />);
    
    fireEvent.change(screen.getByPlaceholderText("Nome do usuário"), { target: { value: "Fulano de Tal" } });
    fireEvent.change(screen.getByPlaceholderText("exemplo@empresa.com"), { target: { value: "fulano@empresa.com" } });
    
    // Senha sem números/maiúsculas
    fireEvent.change(screen.getByPlaceholderText("Senha (mínimo de 12 caracteres)"), { target: { value: "semmaiusculasegura" } });
    
    fireEvent.submit(screen.getByRole("button", { name: "Criar usuário" }));

    expect(screen.getByText("A senha deve conter ao menos uma letra maiúscula, uma minúscula e um número.")).toBeInTheDocument();
  });

  it("calls onClose when clicking outside (on the backdrop)", () => {
    const onCloseMock = vi.fn();
    render(<UserFormModal {...defaultProps} onClose={onCloseMock} />);
    
    // O backdrop é o elemento mais externo (com classe fixed inset-0)
    const backdrop = screen.getByText("Criar Novo Usuário").closest(".fixed");
    expect(backdrop).toBeInTheDocument();
    
    if (backdrop) {
      fireEvent.click(backdrop);
    }
    
    expect(onCloseMock).toHaveBeenCalledTimes(1);
  });

  it("does not call onClose when clicking inside the modal container", () => {
    const onCloseMock = vi.fn();
    render(<UserFormModal {...defaultProps} onClose={onCloseMock} />);
    
    // Clicar no formulário ou no título não deve fechar
    const title = screen.getByText("Criar Novo Usuário");
    fireEvent.click(title);
    
    expect(onCloseMock).not.toHaveBeenCalled();
  });

  it("calls onSubmit with correct data when validations pass", () => {
    const onSubmitMock = vi.fn();
    render(<UserFormModal {...defaultProps} onSubmit={onSubmitMock} />);
    
    fireEvent.change(screen.getByPlaceholderText("Nome do usuário"), { target: { value: "Fulano de Tal" } });
    fireEvent.change(screen.getByPlaceholderText("exemplo@empresa.com"), { target: { value: "fulano@empresa.com" } });
    
    // Senha forte que cumpre os critérios (Score 4)
    // Comprimento >= 12, maiúscula, minúscula, número, sem sequências repetitivas simples
    fireEvent.change(screen.getByPlaceholderText("Senha (mínimo de 12 caracteres)"), { target: { value: "A1b9C2d8E3f7" } });
    fireEvent.change(screen.getByPlaceholderText("Confirmar nova senha"), { target: { value: "A1b9C2d8E3f7" } });
    
    fireEvent.submit(screen.getByRole("button", { name: "Criar usuário" }));

    expect(onSubmitMock).toHaveBeenCalledWith({
      name: "Fulano de Tal",
      email: "fulano@empresa.com",
      password: "A1b9C2d8E3f7",
      roleId: null,
    });
  });
});
