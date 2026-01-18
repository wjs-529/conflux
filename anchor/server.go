package anchor

type StartAnchorArgs struct {
	GuardianURL string
	VeilURL     string
	VeilPort    int
	AnchorToken string
	Portal      bool
}

type CreateTUNArgs struct {
	Ifname string
	MTU    int
}

type LinkWithFileDescriptorArgs struct {
	FileDescriptor int
}

type AnchorRPCServer struct {
	Impl Anchor
}

func (s *AnchorRPCServer) CreateAnchor(args interface{}, resp *string) error {
	err := s.Impl.CreateAnchor()
	if err != nil {
		*resp = err.Error()
	} else {
		*resp = ""
	}
	return nil
}

func (s *AnchorRPCServer) DestroyAnchor(args interface{}, resp *string) error {
	err := s.Impl.DestroyAnchor()
	if err != nil {
		*resp = err.Error()
	} else {
		*resp = ""
	}
	return nil
}

func (s *AnchorRPCServer) StartAnchor(args *StartAnchorArgs, resp *string) error {
	err := s.Impl.StartAnchor(args.GuardianURL, args.VeilURL, args.VeilPort, args.AnchorToken, args.Portal)
	if err != nil {
		*resp = err.Error()
	} else {
		*resp = ""
	}
	return nil
}

func (s *AnchorRPCServer) StopAnchor(args interface{}, resp *string) error {
	err := s.Impl.StopAnchor()
	if err != nil {
		*resp = err.Error()
	} else {
		*resp = ""
	}
	return nil
}

func (s *AnchorRPCServer) CreateTUN(args *CreateTUNArgs, resp *string) error {
	err := s.Impl.CreateTUN(args.Ifname, args.MTU)
	if err != nil {
		*resp = err.Error()
	} else {
		*resp = ""
	}
	return nil
}

func (s *AnchorRPCServer) DestroyTUN(args interface{}, resp *string) error {
	err := s.Impl.DestroyTUN()
	if err != nil {
		*resp = err.Error()
	} else {
		*resp = ""
	}
	return nil
}

func (s *AnchorRPCServer) LinkWithTUN(args interface{}, resp *string) error {
	err := s.Impl.LinkWithTUN()
	if err != nil {
		*resp = err.Error()
	} else {
		*resp = ""
	}
	return nil
}

func (s *AnchorRPCServer) LinkWithFileDescriptor(args *LinkWithFileDescriptorArgs, resp *string) error {
	err := s.Impl.LinkWithFileDescriptor(args.FileDescriptor)
	if err != nil {
		*resp = err.Error()
	} else {
		*resp = ""
	}
	return nil
}

func (s *AnchorRPCServer) GetID(args interface{}, resp *string) error {
	id, err := s.Impl.GetID()
	if err != nil {
		return err
	}
	*resp = id
	return nil
}
